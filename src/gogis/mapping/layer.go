package mapping

import (
	"fmt"
	"gogis/base"
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

func (this *Layer) Draw(canvas *draw.Canvas) int64 {
	canvas.SetStyle(this.Style)

	tr := base.NewTimeRecorder()
	feait := this.feaset.QueryByBounds(canvas.Params.GetBounds())
	tr.Output("query layer " + this.Name)
	objCount := feait.Count()
	fmt.Println("object result: ", objCount)

	forCount := objCount/ONE_DRAW_COUNT + 1
	// 直接绘制
	if forCount == 1 {
		this.drawBatch(feait, 0, canvas)
	} else {
		// 并发绘制
		var wg *sync.WaitGroup = new(sync.WaitGroup)
		for i := 0; i < int(forCount); i++ {
			wg.Add(1)
			go this.goDrawBatch(feait, i*ONE_DRAW_COUNT, canvas, wg)
		}
		wg.Wait()
	}

	tr.Output("draw layer " + this.Name)
	feait.Close()
	return objCount
}

// 批量绘制
func (this *Layer) drawBatch(itr data.FeatureIterator, pos int, canvas *draw.Canvas) {
	// tr := base.NewTimeRecorder()
	features, _, ok := itr.BatchNext(int64(pos), ONE_DRAW_COUNT)
	// tr.Output("feaitr fetch batch")
	if ok {
		for _, v := range features {
			drawGeo, ok := v.Geo.(draw.DrawCanvas)
			if ok {
				drawGeo.Draw(canvas)
			}
		}
	}
	features = features[:0]
	// tr.Output("draw batch")
}

// 并发绘制
func (this *Layer) goDrawBatch(itr data.FeatureIterator, pos int, canvas *draw.Canvas, wg *sync.WaitGroup) {
	defer wg.Done()
	canvasBatch := canvas.Clone()
	this.drawBatch(itr, pos, canvasBatch)
	// tr := base.NewTimeRecorder()
	canvas.DrawImage(canvasBatch.Image(), 0, 0)
	// tr.Output("canvas draw image")
}
