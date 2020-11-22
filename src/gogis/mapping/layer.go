package mapping

import (
	"fmt"
	"gogis/data"
	"gogis/draw"
	"sync"
)

type Layer struct {
	Name   string          // 图层名
	feaset data.Featureset // 干活用的
	Params data.ConnParams `json:"ConnParams"` // 存储和打开地图文档时用的
	Style  draw.Style
}

func NewLayer(feaset data.Featureset) *Layer {
	layer := new(Layer)
	layer.feaset = feaset
	store := feaset.GetStore()
	if store != nil {
		layer.Params = store.GetConnParams()
		layer.Params["name"] = feaset.GetName()
		layer.Name = layer.Params["name"] // 默认图层名 等于 数据集名
	}
	//
	layer.Style = draw.RandStyle()
	fmt.Println("layer style:", layer.Style)

	return layer
}

// 一次性绘制的对象个数
const ONE_DRAW_COUNT = 100000

func (this *Layer) Draw(canvas *draw.Canvas) int {
	// this.Style.LineColor = color.RGBA{255, 2, 2, 255}
	// fmt.Println("layer style:", this.Style)
	canvas.SetStyle(this.Style)

	feait := this.feaset.QueryByBounds(canvas.Params.GetBounds())
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for {
		features, ok := feait.BatchNext(ONE_DRAW_COUNT)
		if ok {
			wg.Add(1)
			go this.drawBatch(features, canvas, wg)
		} else {
			break
		}
	}
	wg.Wait()

	return feait.Count()
}

func (this *Layer) drawBatch(features []data.Feature, canvas *draw.Canvas, wg *sync.WaitGroup) {
	defer wg.Done()
	drawCanvas := canvas.Clone()
	for _, v := range features {
		v.Geo.Draw(drawCanvas)
	}
	canvas.DrawImage(drawCanvas.Image(), 0, 0)
}
