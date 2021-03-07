package mapping

import (
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"runtime"
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
const ONE_DRAW_MIN_COUNT = 1000000
const ONE_DRAW_MAX_COUNT = 20000000

var BATCH_COUNT = 10000000

// 设置默认值，New出来的时候调用
func (this *SimpleTheme) MakeDefault(feaset data.Featureset) {
	this.Style = draw.RandStyle()
}

func (this *SimpleTheme) WhenOpenning() {
}

func getObjCount(count int64) (objCount int) {
	objCount = BATCH_COUNT
	if objCount == 0 {
		objCount = int(count) / (runtime.NumCPU())
		objCount = base.IntMax(objCount, ONE_DRAW_MIN_COUNT)
		objCount = base.IntMin(objCount, ONE_DRAW_MAX_COUNT)
	}
	return
}

func (this *SimpleTheme) Draw(canvas draw.Canvas, feaItr data.FeatureIterator, prjc *base.PrjConvert) int64 {
	canvas.SetStyle(this.Style)

	// tr := base.NewTimeRecorder()
	objCount := getObjCount(feaItr.Count())
	forCount := feaItr.BeforeNext(objCount)
	// fmt.Println("go count:", forCount)
	// tr.Output("query layer " + this.Name + ", object count:" + strconv.Itoa(int(objCount)) + ", go count:" + strconv.Itoa(forCount))

	if forCount == 1 {
		// 直接绘制
		this.drawBatch(feaItr, 0, canvas, prjc)
	} else {
		// 并发绘制
		// var wg *sync.WaitGroup = new(sync.WaitGroup)
		var gm base.GoMax
		gm.Init(runtime.NumCPU())
		for i := 0; i < int(forCount); i++ {
			// wg.Add(1)
			gm.Add()
			go this.goDrawBatch(feaItr, i, canvas, prjc, &gm)
		}
		// wg.Wait()
		gm.Wait()
	}

	// tr.Output("draw layer " + this.Name)
	return feaItr.Count()
}

func (this *SimpleTheme) goDrawBatch(itr data.FeatureIterator, batchNo int, canvas draw.Canvas, prjc *base.PrjConvert, gm *base.GoMax) {
	canvasBatch := canvas.Clone()
	// tr := base.NewTimeRecorder()
	// tr.Output("begion:" + strconv.Itoa(batchNo))
	this.drawBatch(itr, batchNo, canvasBatch, prjc)
	canvas.DrawImage(canvasBatch.GetImage(), 0, 0)
	// tr.Output("end:" + strconv.Itoa(batchNo))
	gm.Done()
}

func (this *SimpleTheme) drawBatch(itr data.FeatureIterator, batchNo int, canvas draw.Canvas, prjc *base.PrjConvert) {
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
