package mapping

import (
	"encoding/json"
	"fmt"
	"gogis/base"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	// "github.com/chai2010/webp"
)

type Map struct {
	Name   string
	Layers []*Layer // 0:最底层，先绘制
	canvas *Canvas  // 画布
	BBox   Rect2D   // 所有数据的边框
}

// 复制一个map对象，用来同一个地图的并发出图
func (this *Map) Copy() (nmap *Map) {
	nmap = new(Map)
	nmap.Layers = this.Layers
	nmap.BBox = this.BBox
	nmap.Name = this.Name
	nmap.canvas = new(Canvas)
	nmap.canvas.params = this.canvas.params
	return
}

// 创建一个新地图，设置地图大小
func NewMap() *Map {
	gmap := new(Map)
	gmap.Name = "未命名地图" + string(time.Now().Unix())
	gmap.canvas = new(Canvas)
	// 新建一个 指定大小的 RGBA位图
	// gmap.canvas.img = image.NewNRGBA(image.Rect(0, 0, dx, dy))
	gmap.BBox.Init() // 初始化bbox
	return gmap
}

// 更改画布尺寸
// func (this *Map) Resize(dx int, dy int) {
// 	if dx != this.canvas.img.Rect.Dx() || dy != this.canvas.img.Rect.Dy() {
// 		this.canvas.img = image.NewNRGBA(image.Rect(0, 0, dx, dy))
// 	}
// }

func (this *Map) RebuildBBox() {
	this.BBox.Init()
	for _, layer := range this.Layers {
		bbox := new(Rect2D)
		bbox.Min.X = layer.Shp.xmin
		bbox.Min.Y = layer.Shp.ymin
		bbox.Max.X = layer.Shp.xmax
		bbox.Max.Y = layer.Shp.ymax
		this.BBox.Union(*bbox)
	}
}

func (this *Map) AddLayer(shp *ShapeFile) {
	layer := NewLayer(shp)
	this.Layers = append(this.Layers, layer)
	bbox := new(Rect2D)
	bbox.Min.X = shp.shpHeader.xmin
	bbox.Min.Y = shp.shpHeader.ymin
	bbox.Max.X = shp.shpHeader.xmax
	bbox.Max.Y = shp.shpHeader.ymax
	// fmt.Println("bbox1:", this.BBox)
	// fmt.Println("bbox2:", bbox)
	this.BBox.Union(*bbox)
	// fmt.Println("bbox3:", this.BBox)
}

// 计算各类参数，为绘制做好准备
func (this *Map) Prepare(dx, dy int) {
	this.canvas.params.Init(this.BBox, dx, dy)
}

// 返回绘制对象的个数
func (this *Map) Draw() int {
	// this.prepare()
	this.canvas.img = image.NewNRGBA(image.Rect(0, 0, this.canvas.params.dx, this.canvas.params.dy))
	// fmt.Println("in draw(), image:", this.canvas.img)
	drawCount := 0
	for _, layer := range this.Layers {
		drawCount += layer.Draw(this.canvas)
	}

	return drawCount
}

func (this *Map) OutputImage() *image.NRGBA {
	return this.canvas.img
}

// todo 输出格式，后面再增加
func (this *Map) Output(filename string, imgType string) {
	imgfile, _ := os.Create(filename)
	defer imgfile.Close()

	switch imgType {
	case "png":
		png.Encode(imgfile, this.canvas.img)
	case "jpg", "jpeg":
		jpeg.Encode(imgfile, this.canvas.img, nil)
	case "webp":
		// webp.Encode(imgfile, this.canvas.img)
	}
	// 以PNG格式保存文件

}

func (this *Map) Save(filename string) {
	data, _ := json.Marshal(*this)
	// fmt.Println("map json:", string(data))
	f, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0766)
	f.Write(data)
	f.Close()
}

func (this *Map) Open(filename string) {
	data, _ := ioutil.ReadFile(filename)
	json.Unmarshal(data, this)
	// fmt.Println("map:", this)
	for _, layer := range this.Layers {
		// fmt.Println("shp file name:", layer.Shp.Filename)
		layer.Shp.Open(layer.Shp.Filename)
		layer.Shp.Load()
		// layer.Shp.BuildVecPyramid()
	}
	this.RebuildBBox()
}

// 缩放，ratio为缩放比率，大于1为放大；小于1为缩小
func (this *Map) Zoom(ratio float64) {
	this.canvas.params.scale *= ratio
}

// 按照固定比例尺来缓存地图瓦片
/* 参考 谷歌 的方式
目录：Level/row/col.png
Level 0：180度
Level 1：90度
Level 2: 45度
......

*/
func (this *Map) Cache(path string) {
	// 第一步，计算缓存的层级范围， 先假设都是经纬度数据
	minLevel, maxLevel := this.calcCacheLevels()
	fmt.Println("min & max Level:", minLevel, maxLevel)

	// 创建根目录
	os.MkdirAll(path, os.ModePerm)

	// 第二步，分层级和范围进行并发生成缓存
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := minLevel; i <= maxLevel; i++ {
		wg.Add(1)
		go this.cacheOneLevel(i, path, wg)
	}
	wg.Wait()
}

// 缓存指定的层级
func (this *Map) cacheOneLevel(level int, path string, wg *sync.WaitGroup) {
	defer wg.Done()

	// 1，创建该层级根目录
	levelPath := filepath.Join(path, strconv.Itoa(level))
	os.MkdirAll(levelPath, os.ModePerm)

	// 2，根据范围计算下一级子目录
	// 先计算，当前层级每个瓦片的边长
	dis := 180.0 / (math.Pow(2, float64(level)))
	// 行的范围
	minRow := int(math.Floor((this.BBox.Min.Y + 90.0) / dis))
	maxRow := int(math.Ceil((this.BBox.Max.Y + 90.0) / dis))
	// 列的范围
	minCol := int(math.Floor((this.BBox.Min.X + 180.0) / dis))
	maxCol := int(math.Ceil((this.BBox.Max.X + 180.0) / dis))

	// 开始生成缓存
	fmt.Println("开始生成第%n层瓦片，行范围：%n-%n，列范围：%n-%n", level, minRow, maxRow, minCol, maxCol)

	// 3，并行生成 某个瓦片
	for i := minRow; i <= maxRow; i++ {
		rowPath := filepath.Join(levelPath, strconv.Itoa(i))
		os.MkdirAll(rowPath, os.ModePerm)
		var wg2 *sync.WaitGroup = new(sync.WaitGroup)
		for j := minCol; j <= maxCol; j++ {
			wg2.Add(1)
			// 具体生成瓦片
			go this.cacheOneTile(level, i, j, rowPath, wg2)
		}
		wg2.Wait()
		deleteEmptyDir(rowPath)
	}

	deleteEmptyDir(levelPath)
}

// 具体生成一个瓦片
func (this *Map) cacheOneTile(level int, row int, col int, path string, wg *sync.WaitGroup) {
	defer wg.Done()

	filename := path + "/" + strconv.Itoa(col) + ".png"

	gmap := this.Copy()
	// 这里关键要把 map要绘制的范围设置对了；即根据 level，row，col来计算bbox
	gmap.BBox = this.calcBBox(level, row, col)
	gmap.Prepare(256, 256)
	drawCount := gmap.Draw()
	// fmt.Println("after draw, tile image:", this.canvas.img)

	// 只有真正绘制对象了，才缓存为文件
	if drawCount > 0 {
		// 还得看一下 image中是否 赋值了
		if gmap.checkDrawn() {
			fmt.Println("tile file name:", filename)
			gmap.Output(filename, "png")
		}
	}
}

// 检查是否真正在image中赋值了
func (this *Map) checkDrawn() bool {
	// return true
	if this.canvas.img.Pix != nil {
		// fmt.Println("tile image:", this.canvas.img)
		for _, v := range this.canvas.img.Pix {
			if v != 0 {
				return true
			}
		}
	}
	return false
}

func (this *Map) calcBBox(level int, row int, col int) (bbox Rect2D) {
	dis := 180.0 / (math.Pow(2, float64(level)))
	bbox.Min.X = float64(col)*dis - 180
	bbox.Max.X = bbox.Min.X + dis
	bbox.Min.Y = float64(row)*dis - 90
	bbox.Max.Y = bbox.Min.Y + dis
	return
}

// 根据bbox和对象数量，计算缓存的最小最大合适层级
// 再小的层级没有必要（图片上的显示范围太小）；再大的层级则瓦片上对象太稀疏
func (this *Map) calcCacheLevels() (minLevel, maxLevel int) {
	geoCount := 0
	for _, layer := range this.Layers {
		geoCount += len(layer.Shp.geometrys)
	}

	return base.CalcMinMaxLevels(this.BBox, geoCount)
}
