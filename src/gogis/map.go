package gogis

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
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
	fmt.Println("bbox1:", this.BBox)
	fmt.Println("bbox2:", bbox)
	this.BBox.Union(*bbox)
	fmt.Println("bbox3:", this.BBox)
}

// 计算各类参数，为绘制做好准备
func (this *Map) Prepare(dx, dy int) {
	this.canvas.params.Init(this.BBox, dx, dy)
}

func (this *Map) Draw() {
	// this.prepare()
	this.canvas.img = image.NewNRGBA(image.Rect(0, 0, this.canvas.params.dx, this.canvas.params.dy))

	for _, layer := range this.Layers {
		layer.Draw(this.canvas)
	}
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
	case "webp":
		// webp.Encode(imgfile, this.canvas.img)
	}
	// 以PNG格式保存文件

}

func (this *Map) Save(filename string) {
	data, _ := json.Marshal(*this)
	fmt.Println("map json:", string(data))
	f, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0766)
	f.Write(data)
	f.Close()
}

func (this *Map) Open(filename string) {
	data, _ := ioutil.ReadFile(filename)
	json.Unmarshal(data, this)
	fmt.Println("map:", this)
	for _, layer := range this.Layers {
		fmt.Println("shp file name:", layer.Shp.Filename)
		layer.Shp.Open(layer.Shp.Filename)
		layer.Shp.Load()
	}
	this.RebuildBBox()
}

// 缩放，ratio为缩放比率，大于1为放大；小于1为缩小
func (this *Map) Zoom(ratio float64) {
	this.canvas.params.scale *= ratio
}
