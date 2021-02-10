package mapping

import (
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"strconv"
	"strings"
)

func init() {
	RegisterTheme(ThemeRange, NewRangeTheme)
}

func NewRangeTheme() Theme {
	return new(RangeTheme)
}

type Range struct {
	Min, Max float64
}

func (this *Range) String() string {
	min := strconv.FormatFloat(this.Min, 'f', -1, 64)
	max := strconv.FormatFloat(this.Max, 'f', -1, 64)
	return min + "-" + max
}

func (this *Range) Parse(value string) {
	values := strings.Split(value, "-")
	this.Min, _ = strconv.ParseFloat(values[0], 64)
	this.Max, _ = strconv.ParseFloat(values[1], 64)
}

// 单值专题图，原则上每个对象都有自己的绘制风格
type RangeTheme struct {
	// 对象和风格一一对应
	ranges []Range
	Styles map[string]draw.Style
	Field  string // 字段名 todo
}

func (this *RangeTheme) GetType() ThemeType {
	return ThemeRange
}

func (this *RangeTheme) WhenOpenning() {
	this.ranges = make([]Range, 0, len(this.Styles))
	for k, _ := range this.Styles {
		var rng Range
		rng.Parse(k)
		this.ranges = append(this.ranges, rng)
	}
}

func (this *RangeTheme) MakeDefault(feaset data.Featureset) {
	// feaset.getf
	this.Field = "gid" // todo
	rangeCount := 10   // 默认为10段
	count := feaset.GetCount()
	this.Styles = make(map[string]draw.Style, rangeCount)
	this.ranges = make([]Range, rangeCount)

	for i := 0; i < rangeCount; i++ {
		var rng Range
		rng.Min = float64(i) * (float64(count) / float64(rangeCount))
		rng.Max = rng.Min + float64(count)/float64(rangeCount)
		this.ranges[i] = rng
		this.Styles[rng.String()] = draw.RandStyle()
	}
}

func (this *RangeTheme) Draw(canvas *draw.Canvas, feait data.FeatureIterator, prjc *base.PrjConvert) int64 {
	objCount := 1000
	forCount := feait.BeforeNext(objCount)

	for i := 0; i < forCount; i++ {
		if feas, ok := feait.BatchNext(i); ok {
			for _, v := range feas {
				drawGeo, ok := v.Geo.(draw.DrawCanvas)
				if ok {
					style := this.getStyle(float64(v.Geo.GetID()))
					canvas.SetStyle(style)
					drawGeo.Draw(canvas)
				}

			}
		}
	}
	return feait.Count()
}

func (this *RangeTheme) getStyle(value float64) draw.Style {
	for _, v := range this.ranges {
		if value >= v.Min && value < v.Max {
			return this.Styles[v.String()]
		}
	}
	return draw.RandStyle()
}
