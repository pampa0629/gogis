package mapping

import (
	"encoding/json"
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"

	"github.com/chai2010/webp"
)

type Map struct {
	Name     string
	filename string       // 保存为map文件的文件名
	Layers   []*Layer     // 0:最底层，先绘制
	canvas   *draw.Canvas // 画布
	BBox     base.Rect2D  // 所有数据的边框
}

// 复制一个map对象，用来同一个地图的并发出图
func (this *Map) Copy() (nmap *Map) {
	nmap = new(Map)
	nmap.Layers = this.Layers
	nmap.BBox = this.BBox
	nmap.Name = this.Name
	nmap.canvas = new(draw.Canvas)
	nmap.canvas.Params = this.canvas.Params
	return
}

// 创建一个新地图，设置地图大小
func NewMap() *Map {
	gmap := new(Map)
	// gmap.Name = "未命名地图" + strconv.FormatInt(time.Now().Unix(), 10)
	gmap.canvas = new(draw.Canvas)
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
		this.BBox.Union(layer.feaset.GetBounds())
	}
}

func (this *Map) AddLayer(feaset data.Featureset) {
	if len(this.Name) == 0 {
		this.Name = feaset.GetName()
	}
	layer := NewLayer(feaset)
	this.Layers = append(this.Layers, layer)
	this.BBox.Union(feaset.GetBounds())
}

// 为绘制做好准备，第一次绘制前必须调用
func (this *Map) Prepare(dx, dy int) {
	this.canvas.Init(this.BBox, dx, dy)
}

// 返回绘制对象的个数
func (this *Map) Draw() (drawCount int64) {
	this.canvas.ClearDC()
	for _, layer := range this.Layers {
		drawCount += layer.Draw(this.canvas)
	}
	return
}

func (this *Map) OutputImage() image.Image {
	return this.canvas.Image()
}

// todo 输出格式，后面再增加
func (this *Map) Output(w io.Writer, imgType string) {
	switch imgType {
	case "png":
		png.Encode(w, this.canvas.Image())
	case "jpg", "jpeg":
		jpeg.Encode(w, this.canvas.Image(), nil)
	case "webp":
		webp.Encode(w, this.canvas.Image(), nil)
	default:
		fmt.Println("不支持的图片格式：", imgType)
	}
}

// 输出到文件
func (this *Map) Output2File(filename string, imgType string) {
	imgfile, _ := os.Create(filename)
	defer imgfile.Close()

	this.Output(imgfile, imgType)
}

// 工作空间文件的保存
func (this *Map) Save(filename string) {
	this.filename = filename
	// 文件类型，应修改为相对路径
	for _, layer := range this.Layers {
		storename := layer.Params["filename"]
		if len(storename) > 0 {
			layer.Params["filename"] = base.GetRelativePath(filename, storename)
		}
	}

	data, _ := json.Marshal(*this)
	// fmt.Println("map json:", string(data))
	f, _ := os.Create(filename)
	f.Write(data)
	f.Close()
}

// 打开工作空间文件
func (this *Map) Open(filename string) {
	this.filename = filename

	mapdata, _ := ioutil.ReadFile(filename)
	json.Unmarshal(mapdata, this)
	// fmt.Println("opened map:", this)
	fmt.Println("open map file:"+this.filename+", layers'count:", len(this.Layers))

	// 通过保存的参数恢复数据集
	for i, layer := range this.Layers {
		store := data.NewDatastore(data.StoreType(layer.Params["type"]))
		if store != nil {
			// 恢复为绝对路径
			storename := layer.Params["filename"]
			if len(storename) > 0 {
				layer.Params["filename"] = base.GetAbsolutePath(filename, storename)
			}
			ok, _ := store.Open(layer.Params)
			if ok {
				layer.feaset, _ = store.GetFeasetByName(layer.Params["name"])
			}
		} else {
			this.Layers[i] = nil // todo 应该提供恢复的机制，而不是简单置零
		}
		// fmt.Println("open map file, layer style:", layer.Style)
	}

	// this.RebuildBBox()
}

func (this *Map) Close() {
	for _, layer := range this.Layers {
		layer.feaset.GetStore().Close() // 数据库先关闭
		layer.feaset.Close()
	}
	this.Layers = this.Layers[:0]
	// this.canvas.ClearDC() todo 清空image才行
}

// 缩放，ratio为缩放比率，大于1为放大；小于1为缩小
func (this *Map) Zoom(ratio float64) {
	this.canvas.Params.Scale *= ratio
}
