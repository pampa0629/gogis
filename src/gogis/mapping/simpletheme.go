package mapping

import (
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"sync"
)

func init() {
	RegisterTheme(ThemeSimple, NewSimpleTheme)
}

func NewSimpleTheme() Theme {
	return new(SimpleTheme)
}

// 简单图层，统一风格
type SimpleTheme struct {
	Style draw.Style
}

func (this *SimpleTheme) GetType() ThemeType {
	return ThemeSimple
}

// 一次性绘制的对象个数
const ONE_DRAW_COUNT = 100000

// 设置默认值，New出来的时候调用
func (this *SimpleTheme) MakeDefault(feaset data.Featureset) {
	this.Style = draw.RandStyle()
}

func (this *SimpleTheme) WhenOpenning() {
}

func (this *SimpleTheme) Draw(canvas *draw.Canvas, feaItr data.FeatureIterator, prjc *base.PrjConvert) int64 {
	canvas.SetStyle(this.Style)

	// tr := base.NewTimeRecorder()
	objCount := feaItr.Count()
	forCount := feaItr.PrepareBatch(ONE_DRAW_COUNT)
	// tr.Output("query layer " + this.Name + ", object count:" + strconv.Itoa(int(objCount)) + ", go count:" + strconv.Itoa(forCount))

	if forCount == 1 {
		// 直接绘制
		this.drawBatch(feaItr, 0, canvas, prjc)
	} else {
		// 并发绘制
		var wg *sync.WaitGroup = new(sync.WaitGroup)
		for i := 0; i < int(forCount); i++ {
			wg.Add(1)
			go this.goDrawBatch(feaItr, i, canvas, prjc, wg)
		}
		wg.Wait()
	}

	// tr.Output("draw layer " + this.Name)
	return objCount
}

func (this *SimpleTheme) goDrawBatch(itr data.FeatureIterator, pos int, canvas *draw.Canvas, prjc *base.PrjConvert, wg *sync.WaitGroup) {
	defer wg.Done()
	canvasBatch := canvas.Clone()
	// tr := base.NewTimeRecorder()
	this.drawBatch(itr, pos, canvasBatch, prjc)
	// tr.Output(strconv.Itoa(pos))
	canvas.DrawImage(canvasBatch.Image(), 0, 0)
}

func (this *SimpleTheme) drawBatch(itr data.FeatureIterator, batchNo int, canvas *draw.Canvas, prjc *base.PrjConvert) {
	features, ok := itr.BatchNext(batchNo)
	if ok {
		for _, v := range features {
			if v.Geo != nil {
				v.Geo.ConvertPrj(prjc)
				drawGeo, ok := v.Geo.(draw.DrawCanvas)
				if ok {
					drawGeo.Draw(canvas)
				}
			}
		}
	}
	features = features[:0]
}
