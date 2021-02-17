package mapping

import (
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"runtime"
)

// 栅格图层类
type RasterLayer struct {
	LayerType LayerType
	Name      string `json:"LayerName"` // 图层名
	Filename  string // 对应的数据源链接
	dt        *data.MosaicRaset
}

func newRasterLayer(raset *data.MosaicRaset) *RasterLayer {
	layer := new(RasterLayer)
	layer.LayerType = LayerRaster
	// 默认图层名 等于 数据集名
	layer.Filename = raset.Filename()
	layer.Name = base.GetTitle(layer.Filename)
	layer.dt = raset
	return layer
}

func (this *RasterLayer) GetBounds() base.Rect2D { // base.Bounds
	return this.dt.Bbox
}

// todo
func (this *RasterLayer) GetProjection() *base.ProjInfo { // 得到投影坐标系，没有返回nil
	return nil
}

func (this *RasterLayer) GetName() string {
	return this.Name
}

func (this *RasterLayer) GetType() LayerType {
	return LayerRaster
}

func (this *RasterLayer) GetConnParams() data.ConnParams {
	params := data.NewConnParams()
	params["filename"] = this.Filename
	return params
}

func (this *RasterLayer) Close() {
	this.dt.Close()
}

func (this *RasterLayer) Draw(canvas *draw.Canvas, proj *base.ProjInfo) int64 {
	// todo 动态投影
	bbox := canvas.Params.GetBounds()
	width, height := canvas.GetSize()
	level, nos := this.dt.Perpare(bbox, width, height)
	var gm base.GoMax
	gm.Init(runtime.NumCPU())
	for _, no := range nos {
		gm.Add()
		go goDraw(canvas, this.dt, level, no, &gm)
	}
	gm.Wait()
	return int64(len(nos))
}

func goDraw(canvas *draw.Canvas, dt *data.MosaicRaset, level, no int, gm *base.GoMax) {
	defer gm.Done()
	w, h := canvas.GetSize()
	img, x, y := dt.GetImage(level, no, canvas.Params.GetBounds(), w, h)
	canvas.DrawImage(img, x, y)
}

// 地图 Save时，内部存储调整为相对路径
func (this *RasterLayer) WhenSaving(mappath string) {
	this.Filename = base.GetRelativePath(mappath, this.Filename)
}

func (this *RasterLayer) WhenOpening(mappath string) {
	this.Filename = base.GetAbsolutePath(mappath, this.Filename)
	this.dt = new(data.MosaicRaset)
	this.dt.Open(this.Filename)
}
