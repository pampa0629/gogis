package mapping

import (
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"strconv"
)

func init() {
	RegisterTheme(ThemeUnique, NewUniqueTheme)
}

func NewUniqueTheme() Theme {
	return new(UniqueTheme)
}

// 单值专题图，原则上每个对象都有自己的绘制风格
type UniqueTheme struct {
	// 对象和风格一一对应
	Styles map[string]draw.Style
	Field  string // 单值的字段名 todo
}

func (this *UniqueTheme) GetType() ThemeType {
	return ThemeUnique
}

func (this *UniqueTheme) MakeDefault(feaset data.Featureset) {
	// feaset.getf
	this.Field = "gid" // todo
	count := feaset.GetCount()
	this.Styles = make(map[string]draw.Style, count)
	bbox := feaset.GetBounds()
	feait := feaset.Query(&data.QueryDef{SpatialObj: &bbox})
	objCount := 1000
	forCount := feait.BeforeNext(objCount)

	for i := 0; i < forCount; i++ {
		if feas, ok := feait.BatchNext(i); ok {
			for _, v := range feas {
				this.Styles[strconv.Itoa(int(v.Geo.GetID()))] = draw.RandStyle()
			}
		}
	}
}

func (this *UniqueTheme) WhenOpenning() {

}

func (this *UniqueTheme) Draw(canvas draw.Canvas, feait data.FeatureIterator, prjc *base.PrjConvert) int64 {
	objCount := 1000
	forCount := feait.BeforeNext(objCount)

	for i := 0; i < forCount; i++ {
		if feas, ok := feait.BatchNext(i); ok {
			for _, v := range feas {
				drawGeo, ok := v.Geo.(draw.DrawCanvas)
				if ok {
					style := this.Styles[strconv.Itoa(int(v.Geo.GetID()))]
					canvas.SetStyle(style)
					drawGeo.Draw(canvas)
				}

			}
		}
	}
	return feait.Count()
}
