package data

import (

	// "gogis/data/memory"
	// "gogis/geometry"

	"gogis/base"
	"image"
	"io/ioutil"
	"math"
	"strconv"
	"strings"

	"github.com/lukeroth/gdal"
)

// Mosaic Raster Set
type MosaicRaset struct {
	Filename string // 配置文件
	bbox     base.Rect2D
	// 0 是原始层，之后是缩略图
	dtss    [][]gdal.Dataset
	bboxess [][]base.Rect2D
	ress    []float64 // 本层的分辨率，一个像素代表的地理范围
}

// 构建镶嵌数据集
func (this *MosaicRaset) Build(path, filename string) {
	// todo
}

func (this *MosaicRaset) Close() {
	for _, dts := range this.dtss {
		for _, dt := range dts {
			dt.Close()
		}
	}
	this.dtss = this.dtss[:0]
	this.bboxess = this.bboxess[:0]
	this.ress = this.ress[:0]
}

func (this *MosaicRaset) Open(filename string) {
	this.bbox.Init()
	this.Filename = filename
	var ovPath string
	this.dtss = make([][]gdal.Dataset, 0)
	this.bboxess = make([][]base.Rect2D, 0)
	this.ress = make([]float64, 0)

	data, _ := ioutil.ReadFile(this.Filename)
	for _, line := range strings.Split(string(data), "\n") {
		// onefile := base.GetAbsolutePath(this.Filename, line)
		line = strings.TrimSpace(line)
		ovPath = line
		this.addDataset(line, 0)
	}
	if len(this.bboxess) >= 1 {
		for _, bbox := range this.bboxess[0] {
			this.bbox = this.bbox.Union(bbox)
		}
	}
	// 看看是否有概视图
	this.openOverviews(ovPath)
}

func (this *MosaicRaset) openOverviews(upPath string) {
	ovPath := base.GetAbsolutePath(upPath, "./Overviews_New_DatasetMosaic/")
	// 搜索里面的所有tif文件
	infos, _ := ioutil.ReadDir(ovPath)
	for _, info := range infos {
		name := info.Name()
		if strings.HasSuffix(name, ".tif") || strings.HasSuffix(name, ".tiff") {
			filename := base.GetAbsolutePath(ovPath+"/", name)
			level := getLevel(name)
			if level > 0 {
				this.addDataset(filename, level)
			}
		}
	}
}

func (this *MosaicRaset) addDataset(filename string, level int) {
	for len(this.dtss) <= level {
		dts := make([]gdal.Dataset, 0)
		this.dtss = append(this.dtss, dts)
		bboxes := make([]base.Rect2D, 0)
		this.bboxess = append(this.bboxess, bboxes)
		this.ress = append(this.ress, 0)
	}

	dt, _ := gdal.Open(filename, 0)
	this.dtss[level] = append(this.dtss[level], dt)
	// 范围
	bbox := this.calcBbox(dt)
	this.bboxess[level] = append(this.bboxess[level], bbox)

	if this.ress[level] == 0 {
		resX := bbox.Dx() / float64(dt.RasterXSize())
		resY := bbox.Dy() / float64(dt.RasterYSize())
		this.ress[level] = math.Max(resX, resY) // 存下比较粗糙的分辨率方向
	}
}

// Ovr_L1_0x0.tif -- > 1
func getLevel(name string) int {
	names := strings.Split(name, "_")
	if len(names) >= 2 {
		str := strings.Replace(names[1], "L", "", 1)
		level, _ := strconv.Atoi(str)
		return level
	}
	return 0
}

// GeoTransform[0],GeoTransform[3]  左上角位置
// GeoTransform[1]是像元宽度
// GeoTransform[5]是像元高度
// 如果影像是指北的,GeoTransform[2]和GeoTransform[4]这两个参数的值为0。
func (this *MosaicRaset) calcBbox(dt gdal.Dataset) (bbox base.Rect2D) {
	tf := dt.GeoTransform()
	xs := dt.RasterXSize()
	ys := dt.RasterYSize()
	bbox.Min.X = tf[0]
	bbox.Min.Y = tf[3] + float64(ys)*tf[5]
	bbox.Max.X = tf[0] + float64(xs)*tf[1]
	bbox.Max.Y = tf[3]
	return bbox
}

func (this *MosaicRaset) GetBounds() base.Rect2D {
	return this.bbox
}

// 返回有哪些 dt 需要
func (this *MosaicRaset) Perpare(bbox base.Rect2D, width, height int) (level int, nos []int) {
	resX := bbox.Dx() / float64(width)
	resY := bbox.Dy() / float64(height)
	res := math.Min(resX, resY) // 用比较细致的分辨率进行层级确定
	for i := len(this.ress) - 1; i >= 0; i-- {
		// 获取比要求更细致的层级
		if this.ress[i] < res {
			level = i
			break
		}
	}

	for i, _ := range this.dtss[level] {
		interBbox := this.bboxess[level][i].Intersects(bbox)
		if interBbox.IsValid() {
			nos = append(nos, i)
		}
	}
	return
}

// 根据地理范围，获取要绘制的image
// bbox:地理范围；width, height：canvas的尺寸
// 返回的x、y是 img在canvas的位置
func (this *MosaicRaset) GetImage(level, no int, bbox base.Rect2D, width, height int) (img *image.RGBA, x, y int) {
	dt := this.dtss[level][no]
	interBbox := this.bboxess[level][no].Intersects(bbox)

	if interBbox.IsValid() {
		// 先计算新的img的尺寸
		dx := int(interBbox.Dx() / bbox.Dx() * float64(width))
		dy := int(interBbox.Dy() / bbox.Dy() * float64(height))
		img = image.NewRGBA(image.Rect(0, 0, dx, dy))
		// 再计算新img在canvas的位置
		x = int((interBbox.Min.X - bbox.Min.X) / bbox.Dx() * float64(width))
		y = int((bbox.Max.Y - interBbox.Max.Y) / bbox.Dy() * float64(height))
		// 最后从dt中取数据，并绘制到img中
		renderImage(dt, interBbox, this.bboxess[level][no], img)
	}
	// }
	return
}

// interBbox:相交部分矩形的地理范围
// bbox：dt的地理范围
// img：相交部分要绘制的image
func renderImage(dt gdal.Dataset, interBbox, bbox base.Rect2D, img *image.RGBA) {
	if dt.RasterCount() < 3 {
		// todo
		return
	}
	// 内部分两个环节：1）取对应的数据；2）填入img中

	band1 := dt.RasterBand(1)
	band2 := dt.RasterBand(2)
	band3 := dt.RasterBand(3)

	// 计算取数据的起止点（像素）
	sx, sy := band1.XSize(), band2.YSize()
	var start, end image.Point
	start.X = int((interBbox.Min.X - bbox.Min.X) / bbox.Dx() * float64(sx))
	start.Y = int((bbox.Max.Y - interBbox.Max.Y) / bbox.Dy() * float64(sy))
	end.X = int((interBbox.Max.X - bbox.Min.X) / bbox.Dx() * float64(sx))
	end.Y = int((bbox.Max.Y - interBbox.Min.Y) / bbox.Dy() * float64(sy))
	// 看看取多少行列
	sizeX := end.X - start.X
	sizeY := end.Y - start.Y

	// 构造内存块
	imgSizeX := img.Rect.Dx()
	imgSizeY := img.Rect.Dy()
	imageSize := imgSizeX * imgSizeY
	data1 := make([]uint8, imageSize)
	data2 := make([]uint8, imageSize)
	data3 := make([]uint8, imageSize)

	// gdal 牛逼，IO直接取数据，自动判断金字塔和进行压缩
	band1.IO(gdal.Read, start.X, start.Y, sizeX, sizeY, data1, imgSizeX, imgSizeY, 1, 0)
	band2.IO(gdal.Read, start.X, start.Y, sizeX, sizeY, data2, imgSizeX, imgSizeY, 1, 0)
	band3.IO(gdal.Read, start.X, start.Y, sizeX, sizeY, data3, imgSizeX, imgSizeY, 1, 0)

	data := make([]uint8, imageSize*4)
	for i := 0; i < imageSize; i++ {
		data[4*i+0] = data1[i]
		data[4*i+1] = data2[i]
		data[4*i+2] = data3[i]
		data[4*i+3] = 255
	}
	copy(img.Pix, data)
	return
}
