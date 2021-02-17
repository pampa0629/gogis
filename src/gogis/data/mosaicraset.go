package data

import (

	// "gogis/data/memory"
	// "gogis/geometry"

	"encoding/json"
	"gogis/base"
	"gogis/index"
	"image"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/lukeroth/gdal"
)

// Mosaic Raster Set
type MosaicRaset struct {
	// lock     sync.Mutex
	filename string // 配置文件
	Bbox     base.Rect2D
	// 0 是原始层，之后是缩略图
	dtss    [][]*gdal.Dataset
	Namess  [][]string
	Bboxess [][]base.Rect2D
	Ress    []float64 // 本层的分辨率，一个像素代表的地理范围
}

// 构建镶嵌数据集；包括创建金字塔、缩略图和保存gmr配置文件。
// 金字塔（overview）或缩略图若已存在，不会重复创建
func (this *MosaicRaset) Build(path, filename string) {
	// tr := base.NewTimeRecorder()
	// 1）搜索所有可识别的影像文件，计算范围和分辨率
	filepath.Walk(path, this.WalkDS)
	// tr.Output("walk")

	// 2）并行创建金字塔
	this.buildOverviews()
	// tr.Output("buildOverviews")

	// 3）创建缩略图
	this.buildThumbnails(path)
	// tr.Output("buildThumbnails")

	// 4）保存为 gmr文件
	this.Save(filename)
	// tr.Output("Save")
}

// 创建缩略图
func (this *MosaicRaset) buildThumbnails(path string) {
	tnPath := base.GetAbsolutePath(path+"/", "./Thumbnails")
	os.MkdirAll(tnPath, os.ModePerm)

	// 思路：采用R-tree索引，先构建索引树，然后从子节点开始创建缩略图
	var rtree index.RTreeIndex
	rtree.Init(this.Bbox, int64(len(this.dtss[0])))
	index.RTREE_OBJ_COUNT = 8 // 多少合适呢？
	for i, v := range this.Bboxess[0] {
		rtree.AddOne(v, int64(i))
	}
	maxLevel := rtree.Level()
	this.buildTnNode(&rtree.RTreeNode, maxLevel, tnPath+"/")
}

func (this *MosaicRaset) buildTnNode(node *index.RTreeNode, maxLevel int, tnPath string) (no int) {
	rLevel, isLeaf, bbox, nodes, ids := node.GetData()
	level := maxLevel - rLevel
	if isLeaf {
		no = this.buildTnLeaf(level, ids, bbox, tnPath)
	} else {
		nos := make([]int64, len(nodes))
		for i, node := range nodes {
			nos[i] = int64(this.buildTnNode(node, maxLevel, tnPath))
		}
		// 中间节点要根据下级节点的结果来生成
		no = this.buildTnLeaf(level, nos, bbox, tnPath)
	}
	return
}

// tnPath: 缩略图的总目录
// level：缩略图的层级
// ids：下一级数据文件在 dtss[level-1]中的index
func (this *MosaicRaset) buildTnLeaf(level int, ids []int64, bbox base.Rect2D, tnPath string) int {
	levelPath := tnPath + strconv.Itoa(level) + "/"
	os.MkdirAll(levelPath, os.ModePerm)

	// 每一层的分辨率至少模糊N倍，并通过分辨率来反算size
	res := this.Ress[level-1] * 3
	sizeX := bbox.Dx() / res
	sizeY := bbox.Dy() / res
	// 设置一个最大尺寸
	maxSize := 4096.0
	for sizeX > maxSize || sizeY > maxSize {
		sizeX /= 2
		sizeY /= 2
	}

	// 创建数据文件
	no := 0
	if len(this.dtss) > level {
		no = len(this.dtss[level])
	}
	filename := levelPath + "tn_" + strconv.Itoa(no) + ".tif"

	// 不存在，才创建缩略图tiff
	if !base.IsExist(filename) {
		drive, _ := gdal.GetDriverByName("GTiff")
		dt := drive.Create(filename, int(sizeX), int(sizeY), 3, gdal.Byte, []string{})

		// 读取并写入数据
		for _, id := range ids {
			data, x, y, dx, dy := this.GetData(level-1, int(id), bbox, int(sizeX), int(sizeY), false)
			if data != nil {
				dt.IO(gdal.Write, x, y, dx, dy, data, dx, dy, 3, []int{1, 2, 3}, 3, dx*3, 1)
			}
		}

		// 设置地理范围
		var tfs [6]float64
		tfs[0] = bbox.Min.X
		tfs[3] = bbox.Max.Y
		tfs[1] = bbox.Dx() / sizeX
		tfs[5] = -1 * bbox.Dy() / sizeY
		dt.SetGeoTransform(tfs)
		dt.Close()
	}

	// 加入，并创建金字塔
	this.addDataset(filename, level, no)
	this.buildOverview(this.dtss[level][no], level, no, nil)
	return no
}

// 创建金字塔文件
func (this *MosaicRaset) buildOverviews() {
	var gm base.GoMax
	gm.Init(runtime.NumCPU())
	for i, dts := range this.dtss {
		for j, dt := range dts {
			gm.Add()
			go this.buildOverview(dt, i, j, &gm)
		}
	}
	gm.Wait()
	this.computeBounds()
}

func (this *MosaicRaset) buildOverview(dt *gdal.Dataset, level, no int, gm *base.GoMax) {
	if gm != nil {
		defer gm.Done()
	}

	// todo 还有其它类型的金字塔文件 .aux .rrd
	if !base.IsExist(this.Namess[level][no] + ".ovr") {
		size := base.IntMax(dt.RasterXSize(), dt.RasterYSize())
		var ovList, bands []int
		ratio := 2
		for size > 256 {
			ovList = append(ovList, ratio)
			ratio *= 2
			size /= 2
		}
		for i := 1; i <= dt.RasterCount(); i++ {
			bands = append(bands, i)
		}
		dt.BuildOverviews("NEAREST", len(ovList), ovList, len(bands), bands,
			func(complete float64, message string, progressArg interface{}) int { return 1 }, nil)
	}
}

func (this *MosaicRaset) WalkDS(filepath string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if f == nil || f.IsDir() || strings.Index(filepath, "Thumbnails") > 0 {
		return nil
	}
	// todo 未来应识别更多数据文件
	ext := path.Ext(filepath)
	switch ext {
	case ".tif", ".tiff":
		this.addDataset(filepath, 0, -1)
	}
	return nil
}

// =============================================== //

func (this *MosaicRaset) Close() {
	for _, dts := range this.dtss {
		for _, dt := range dts {
			if dt != nil {
				dt.Close()
			}
		}
	}
	this.dtss = this.dtss[:0]
	this.Bboxess = this.Bboxess[:0]
	this.Ress = this.Ress[:0]
}

func (this *MosaicRaset) Save(filename string) {
	// namess路径需要改为相对路径
	for i, names := range this.Namess {
		for j, name := range names {
			this.Namess[i][j] = base.GetRelativePath(filename, name)
		}
	}
	data, err := json.MarshalIndent(*this, "", "   ")
	base.PrintError("save mosaic", err)

	this.filename = strings.TrimSuffix(filename, path.Ext(filename)) + ".gmr"
	f, _ := os.Create(this.filename)
	f.Write(data)
	f.Close()
}

// txt是SuperMap镶嵌数据集导出的清单文件
func (this *MosaicRaset) openTxt(filename string) {
	this.filename = filename
	var ovPath string
	this.dtss = make([][]*gdal.Dataset, 0)
	this.Bboxess = make([][]base.Rect2D, 0)
	this.Ress = make([]float64, 0)
	this.Namess = make([][]string, 0)

	data, _ := ioutil.ReadFile(this.filename)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			ovPath = line
			this.addDataset(line, 0, -1)
		}
	}
	this.computeBounds()

	// 看看是否有缩略图
	this.openThumbnail(filepath.Dir(ovPath) + "/")
}

func (this *MosaicRaset) computeBounds() {
	this.Bbox.Init()
	if len(this.Bboxess) >= 1 {
		for _, bbox := range this.Bboxess[0] {
			this.Bbox = this.Bbox.Union(bbox)
		}
	}
}

// gmr: gogis mosaic raster
func (this *MosaicRaset) openGmr(filename string) {
	this.filename = filename
	mapdata, _ := ioutil.ReadFile(filename)
	json.Unmarshal(mapdata, this)
	// 路径需要改为绝对路径，方便后续读取文件
	for i, names := range this.Namess {
		for j, name := range names {
			this.Namess[i][j] = base.GetAbsolutePath(filename, name)
			// this.addDataset(this.Namess[i][j], i, j)
		}
	}
}

func (this *MosaicRaset) Filename() string {
	return this.filename
}

func (this *MosaicRaset) Open(filename string) {
	if strings.HasSuffix(filename, ".gmr") {
		this.openGmr(filename)
	} else {
		this.openTxt(filename)
	}
}

// 查找缩略图信息
func (this *MosaicRaset) openThumbnail(upPath string) {
	// 先确定缩略图所在的目录
	var tnPath string
	tnPaths, _ := ioutil.ReadDir(upPath)
	for _, v := range tnPaths {
		if v.IsDir() && v.Name()[0:10] == "Overviews_" {
			tnPath = v.Name()
		}
	}

	// 换为绝对路径
	tnPath = base.GetAbsolutePath(upPath, tnPath)
	// 搜索里面的所有tif文件
	infos, _ := ioutil.ReadDir(tnPath)
	for _, info := range infos {
		name := info.Name()
		if strings.HasSuffix(name, ".tif") || strings.HasSuffix(name, ".tiff") {
			filename := base.GetAbsolutePath(tnPath+"/", name)
			level := getLevel(name)
			if level > 0 {
				this.addDataset(filename, level, -1)
			}
		}
	}

}

func (this *MosaicRaset) addDataset(filename string, level, no int) {
	// this.lock.Lock()
	for len(this.dtss) <= level {
		dts := make([]*gdal.Dataset, 0)
		this.dtss = append(this.dtss, dts)
	}
	for len(this.Namess) <= level {
		names := make([]string, 0)
		this.Namess = append(this.Namess, names)
		bboxes := make([]base.Rect2D, 0)
		this.Bboxess = append(this.Bboxess, bboxes)
		this.Ress = append(this.Ress, 0)
	}

	// < 0 表示在最后增加一个
	if no < 0 {
		no = len(this.dtss[level])
	}
	// 长度自动扩展到no
	for len(this.dtss[level]) <= no {
		this.dtss[level] = append(this.dtss[level], nil)
	}
	for len(this.Namess[level]) <= no {
		this.Namess[level] = append(this.Namess[level], "")
		var bbox base.Rect2D
		bbox.Init()
		this.Bboxess[level] = append(this.Bboxess[level], bbox)
	}

	// dt
	dt, _ := gdal.Open(filename, 0)
	this.dtss[level][no] = &dt
	// name
	this.Namess[level][no] = filename
	// 范围
	bbox := this.calcBbox(dt)
	this.Bboxess[level][no] = bbox
	// res
	if this.Ress[level] == 0 {
		resX := bbox.Dx() / float64(dt.RasterXSize())
		resY := bbox.Dy() / float64(dt.RasterYSize())
		this.Ress[level] = math.Max(resX, resY) // 存下比较粗糙的分辨率方向
	}
	// this.lock.Unlock()
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
	return this.Bbox
}

// 分辨率,值越小越精细
func (this *MosaicRaset) GetResolution() (res float64) {
	dt := this.getDataset(0, 0)
	if dt != nil {
		resX := this.Bboxess[0][0].Dx() / float64(dt.RasterXSize())
		resY := this.Bboxess[0][0].Dy() / float64(dt.RasterYSize())
		res = math.Min(resX, resY)
	}
	return
}

// =============================================== //

// 返回有哪些 dt 需要
func (this *MosaicRaset) Perpare(bbox base.Rect2D, width, height int) (level int, nos []int) {
	resX := bbox.Dx() / float64(width)
	resY := bbox.Dy() / float64(height)
	res := math.Min(resX, resY) // 用比较细致的分辨率进行层级确定
	for i := len(this.Ress) - 1; i >= 0; i-- {
		// 获取比要求更细致的层级
		if this.Ress[i] < res {
			level = i
			break
		}
	}

	for i, _ := range this.Bboxess[level] {
		interBbox := this.Bboxess[level][i].Intersects(bbox)
		if interBbox.IsValid() {
			nos = append(nos, i)
		}
	}
	return
}

func (this *MosaicRaset) getDataset(level, no int) *gdal.Dataset {
	if level >= len(this.dtss) || no >= len(this.dtss[level]) || this.dtss[level][no] == nil {
		this.addDataset(this.Namess[level][no], level, no)
	}
	return this.dtss[level][no]
}

// 根据地理范围，获取要绘制的image
// bbox:地理范围；width, height：canvas的尺寸
// 返回的x、y是 img在canvas的位置
func (this *MosaicRaset) GetImage(level, no int, bbox base.Rect2D, width, height int) (img *image.RGBA, x, y int) {
	data, x, y, dx, dy := this.GetData(level, no, bbox, width, height, true)
	img = &image.RGBA{Pix: data, Stride: 4 * dx, Rect: image.Rect(0, 0, dx, dy)}
	return img, x, y
}

// 根据地理范围，获取要绘制的数据
// bbox:地理范围；
// width, height：bbox代表的物理尺寸
// alpha：是否需要带有alpha值
// 返回值：x/y 是data在外部的位置，dx/dy是数据的size
func (this *MosaicRaset) GetData(level, no int, bbox base.Rect2D, width, height int, alpha bool) (data []uint8, x, y, dx, dy int) {
	// gdal不支持并发读取，为了支持协程并发，不得已重新Open
	// dt := this.getDataset(level, no)
	dt, _ := gdal.Open(this.Namess[level][no], gdal.ReadOnly)

	interBbox := this.Bboxess[level][no].Intersects(bbox)
	if interBbox.IsValid() {
		// 先计算要输出的数据的尺寸
		dx = base.Round(interBbox.Dx() / bbox.Dx() * float64(width))
		dy = base.Round(interBbox.Dy() / bbox.Dy() * float64(height))
		// 再计算数据在外部的位置
		x = base.Round((interBbox.Min.X - bbox.Min.X) / bbox.Dx() * float64(width))
		y = base.Round((bbox.Max.Y - interBbox.Max.Y) / bbox.Dy() * float64(height))
		// 最后从dt中取数据
		if dx > 0 && dy > 0 {
			data = this.getData(&dt, interBbox, this.Bboxess[level][no], dx, dy, alpha)
		}
	}
	dt.Close()
	return
}

// interBbox:相交部分矩形的地理范围
// bbox：dt的地理范围
// outSizeX, outSizeY：相交部分输出的物理大小
func (this *MosaicRaset) getData(dt *gdal.Dataset, interBbox, bbox base.Rect2D, outSizeX, outSizeY int, alpha bool) []uint8 {
	if dt.RasterCount() < 3 {
		// todo
		return nil
	}
	// 计算取数据的起止点（像素）
	sx, sy := dt.RasterXSize(), dt.RasterYSize()
	var start, end image.Point
	start.X = base.Round((interBbox.Min.X - bbox.Min.X) / bbox.Dx() * float64(sx))
	start.Y = base.Round((bbox.Max.Y - interBbox.Max.Y) / bbox.Dy() * float64(sy))
	end.X = base.Round((interBbox.Max.X - bbox.Min.X) / bbox.Dx() * float64(sx))
	end.Y = base.Round((bbox.Max.Y - interBbox.Min.Y) / bbox.Dy() * float64(sy))
	// 看看取多少行列
	sizeX := end.X - start.X
	sizeY := end.Y - start.Y

	// 构造内存块
	outSize := outSizeX * outSizeY
	bands := 3
	if alpha {
		bands = 4
	}
	data := make([]uint8, outSize*bands)

	// gdal 牛逼，IO直接取数据，自动判断金字塔和进行数据尺度变换
	/* nPixelSpace，nLineSpace与nBandSpace
	如要想将JPG图像按照RGBRGBRGB……（这里的RGB仅代表一个像素，一共有width*height对RGB），
	首先看到R像素与下一个R元素隔了3个像素，因此nPixelSpace的值为sizeof(DataType)*3；
	第一行与下一行之间隔了原来的两行数据，因此nLineSpace为sizeof(DataType)*width*3；
	第一波段与下一波段之间的距离即为一个像素，即为sizeof(DataType)。
	当数据类型DataType为unsigned char 时，sizeof(unsigned char)为1，sizeof(unsigned short)为2*/
	// fmt.Println(start.X, start.Y, sizeX, sizeY,
	// 	outSizeX, outSizeY, bands, outSizeX*bands, 1)
	dt.IO(gdal.Read, start.X, start.Y, sizeX, sizeY,
		data, outSizeX, outSizeY, 3, []int{1, 2, 3}, bands, outSizeX*bands, 1)

	// A的位置要记得填充255（未来可用设置为图层透明度）
	if alpha {
		for i := 0; i < outSize; i++ {
			value := int(data[4*i+0]) + int(data[4*i+1]) + int(data[4*i+2])
			if value == 0 {
				data[4*i+3] = 0 // 无值的地方，alpha设置为0
			} else {
				data[4*i+3] = 255
			}
		}
	}
	return data
}
