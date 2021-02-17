package mapping

import (
	"encoding/json"
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"image"
	"io/ioutil"
	"os"

	"github.com/tidwall/mvt"
)

type Map struct {
	Name     string `json:"MapName"`
	filename string // 保存为map文件的文件名

	// 屏幕图层，todo
	trackLayer TrackLayer // 跟踪图层，用来绘制被选中的临时对象，不做保存
	Layers     Layers     // 0:最底层，先绘制
	// FeaLayers []*FeatureLayer `json:"FeatureLayers"` // 0:最底层，先绘制
	// RasLayers []*RasterLayer  `json:"RasterLayers"`  // todo 与layers合并

	BBox          base.Rect2D    // 所有数据的边框
	IsDynamicProj bool           `json:"Dynamic Projection"` // 是否支持动态投影
	Proj          *base.ProjInfo `json:"Coordinate System"`

	canvas *draw.Canvas // 画布
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
	gmap.trackLayer.style = draw.HilightStyle()
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
		bbox := layer.GetBounds()
		if this.IsDynamicProj {
			prjc := base.NewPrjConvert(layer.GetProjection(), this.Proj)
			if prjc != nil {
				bbox.Min = prjc.DoPnt(bbox.Min)
				bbox.Max = prjc.DoPnt(bbox.Max)
			}
		}
		this.BBox = this.BBox.Union(bbox)
	}
}

func (this *Map) addLayer(layer Layer) {
	this.Layers = append(this.Layers, layer)
	// todo 设置和开启动态投影时，map的bbos应该发生变化
	this.BBox = this.BBox.Union(layer.GetBounds())
	if len(this.Name) == 0 {
		this.Name = layer.GetName()
	}
	if this.Proj == nil {
		// 自己若没有设置投影系统，则取图层的
		this.Proj = layer.GetProjection()
	}
}

func (this *Map) AddRasterLayer(raset *data.MosaicRaset) {
	layer := newRasterLayer(raset)
	this.addLayer(layer)
}

func (this *Map) AddFeatureLayer(feaset data.Featureset, theme Theme) {
	layer := newFeatureLayer(feaset, theme)
	this.addLayer(layer)
}

// 为绘制做好准备，第一次绘制前必须调用
func (this *Map) Prepare(dx, dy int) {
	// this.canvas.ClearDC()
	this.canvas.Init(this.BBox, dx, dy)
}

func (this *Map) PrepareImage(img *image.RGBA) {
	// this.canvas.ClearDC()
	this.canvas.InitFromImage(this.BBox, img)
}

// 选择，如点击、拉框、多边形等；操作后，被选中的对象放入track layer中
// todo 选择方式包括：包括质心、包含、相交、bbox相交、选中（仅支持点对象）
// todo 暂时只支持obj为矩形； obj 应为 屏幕坐标（而非地图坐标）
func (this *Map) Select(obj interface{}) {
	// 先清空之前的
	this.trackLayer.geos = this.trackLayer.geos[:0]

	for _, layer := range this.Layers {
		if layer.GetType() == LayerFeature {
			feaLayer := layer.(*FeatureLayer)
			// todo 这里要判断图层是否可被选择
			this.trackLayer.geos = append(this.trackLayer.geos, feaLayer.Select(obj)...)
		}
	}
}

// 设置是否动态投影
func (this *Map) SetDynamicProj(isDynamicProj bool) {
	if this.IsDynamicProj != isDynamicProj {
		this.IsDynamicProj = isDynamicProj
		// 重新计算和设置bbox
		this.RebuildBBox()
		width, height := this.canvas.GetSize()
		this.canvas.Params.Init(this.BBox, width, height)
	}
}

// 返回绘制对象的个数
func (this *Map) Draw() (drawCount int64) {
	this.canvas.ClearDC()
	destPrj := this.Proj
	if !this.IsDynamicProj {
		destPrj = nil
	}
	for _, layer := range this.Layers {
		drawCount += layer.Draw(this.canvas, destPrj)
	}
	this.trackLayer.Draw(this.canvas, destPrj)
	return
}

// 输出mvt瓦片数据
func (this *Map) OutputMvt() ([]byte, int64) {
	var count int64
	var tile mvt.Tile
	for _, layer := range this.Layers {
		if layer.GetType() == LayerFeature {
			feaLayer := layer.(*FeatureLayer)
			l := tile.AddLayer(layer.GetName())
			count += feaLayer.OutputMvt(l, this.canvas)
		}
	}
	return tile.Render(), count
}

func (this *Map) OutputImage() image.Image {
	return this.canvas.Image()
}

func (this *Map) Output2Bytes(mapType draw.MapType) []byte {
	return mapType.OutputImg2Bytes(this.canvas.Image())
}

// 输出到文件
func (this *Map) Output2File(filename string, mapType draw.MapType) {
	imgfile, _ := os.Create(filename)
	defer imgfile.Close()
	mapType.OutputImg(imgfile, this.canvas.Image())
}

// 工作空间文件的保存
func (this *Map) Save(filename string) {
	this.filename = filename
	// 文件类型，应修改为相对路径
	for _, layer := range this.Layers {
		layer.WhenSaving(filename)
	}

	data, err := json.MarshalIndent(*this, "", "   ")
	base.PrintError("Map Save", err)

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
	for _, layer := range this.Layers {
		layer.WhenOpening(filename)
	}
}

func (this *Map) Close() {
	for _, layer := range this.Layers {
		layer.Close()
	}
	this.Layers = this.Layers[:0]
}

// 缩放，ratio为缩放比率，大于1为放大；小于1为缩小
func (this *Map) Zoom(ratio float64) {
	this.canvas.Params.Scale *= ratio
}

//  todo
func (this *Map) PanMap(dx, dy float64) {
	this.canvas.Params.MapCenter.X -= dx
	this.canvas.Params.MapCenter.Y -= dy
}
