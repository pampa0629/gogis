package mapping

import (
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"runtime"
)

// 栅格图层类
type RasterLayer struct {
	Name string `json:"LayerName"` // 图层名
	// feaset data.Featureset // 数据来源
	Params data.ConnParams `json:"ConnParams"` // 存储和打开地图文档时用的数据连接信息
	dt     data.MosaicRaset
	// Type   ThemeType       `json:"ThemeType"`
	// theme  Theme           // 专题风格
	// Object interface{}     `json:"Theme"` // 好一招狸猫换太子
}

func newRasterLayer(raset data.MosaicRaset) *RasterLayer {
	layer := new(RasterLayer)
	// 默认图层名 等于 数据集名
	layer.Name = base.GetTitle(raset.Filename)
	layer.dt = raset
	return layer
}

func (this *RasterLayer) Draw(canvas *draw.Canvas, proj *base.ProjInfo) int64 {
	// todo 动态投影
	bbox := canvas.Params.GetBounds()
	width, height := canvas.GetSize()
	level, nos := this.dt.Perpare(bbox, width, height)
	// var wg sync.WaitGroup
	var gm base.GoMax
	gm.Init(runtime.NumCPU())
	for _, no := range nos {
		gm.Add()
		go goDraw(canvas, this.dt, level, no, &gm)
	}
	gm.Wait()
	return int64(len(nos))
}

func goDraw(canvas *draw.Canvas, dt data.MosaicRaset, level, no int, gm *base.GoMax) {
	defer gm.Done()
	w, h := canvas.GetSize()
	img, x, y := dt.GetImage(level, no, canvas.Params.GetBounds(), w, h)
	canvas.DrawImage(img, x, y)
}
