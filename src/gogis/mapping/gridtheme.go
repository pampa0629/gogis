package mapping

import (
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"gogis/geometry"
	"image/color"
	"strconv"
)

func init() {
	RegisterTheme(ThemeGrid, NewGridTheme)
}

func NewGridTheme() Theme {
	return new(GridTheme)
}

// 格网聚合图
// todo 当前只支持es引擎
type GridTheme struct {
	// 颜色表
	Colors draw.ColorTable `json:"ColorTable"`
}

func (this *GridTheme) GetType() ThemeType {
	return ThemeGrid
}

func (this *GridTheme) MakeDefault(feaset data.Featureset) {
	this.Colors.Make(color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255})
}

func (this *GridTheme) WhenOpenning() {

}

// 计算合适的 geohash精度值
func calcGridPrecision(bbox base.Rect2D, width, height int) (precision int) {
	pixel := 120 * 90
	area := (bbox.Area() * float64(pixel)) / float64(width*height)
	global := 180.0 * 360.0
	for area < global {
		area *= 32 // 多一位，面积差32倍
		precision++
	}
	return
}

func (this *GridTheme) Draw(canvas *draw.Canvas, feaItr data.FeatureIterator, prjc *base.PrjConvert) int64 {
	// 文字颜色
	canvas.SetTextColor(color.RGBA{0, 255, 0, 255})
	var style draw.Style

	feait := feaItr.(*data.EsFeaItr)
	width, height := canvas.GetSize()
	precision := calcGridPrecision(canvas.Params.GetBounds(), width, height)
	bboxes, counts := feait.AggGrids(precision)
	max, min := base.GetExtreme(counts)

	var geo geometry.GeoPolygon
	for i, bbox := range bboxes {
		geo.Make(bbox)
		// 先确定颜色，再绘制bbox
		style.FillColor = draw.Color(this.Colors.GetColor(counts[i], max, min))
		canvas.SetStyle(style)
		geo.Draw(canvas)

		// 绘制文字
		pnt := canvas.Params.Forward(bbox.Center())
		canvas.DrawString(strconv.Itoa(counts[i]), pnt.X, pnt.Y)
	}

	return int64(len(counts))
}
