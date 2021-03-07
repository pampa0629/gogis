package draw

import (
	"gogis/base"
	"math"
)

// 通用的坐标转化参数
type CoordParams struct {
	Scale     float64      // 图片距离/地图距离
	MapCenter base.Point2D // 地图中心点
	dx, dy    float32      // 图片的宽度和高度的一半
	yup       bool         // Y 轴是否向上
	center    Point        // 坐标轴中心的位置，默认为[0,0]
}

// 根据地图的box和要绘制的大小，来初始化参数
func (this *CoordParams) Init(bbox base.Rect2D, width, height float32, yup bool, center Point) {
	scaleX := (float64)(width) / bbox.Dx()
	scaleY := (float64)(height) / bbox.Dy()
	this.Scale = math.Min(scaleX, scaleY)

	this.MapCenter = bbox.Center()
	this.dx = width / 2.0
	this.dy = height / 2.0
	this.yup = yup
	this.center = center
}

// 得到当前地图的范围
func (this *CoordParams) GetBounds() (bbox base.Rect2D) {
	dx := float64(this.dx) / this.Scale
	dy := float64(this.dy) / this.Scale
	bbox.Min.X = this.MapCenter.X - dx
	bbox.Min.Y = this.MapCenter.Y - dy
	bbox.Max.X = this.MapCenter.X + dx
	bbox.Max.Y = this.MapCenter.Y + dy
	return
}

func (this *CoordParams) Zoom(ratio float64) {
	this.Scale *= ratio
}

// todo
func (this *CoordParams) Zoom2(ratio float64, x, y int) {
}

// todo
func (this *CoordParams) Pan(dx, dy int) {

}

// 正向转化一个点坐标：从地图坐标变为图片坐标
func (this *CoordParams) Forward(pnt base.Point2D) (drawPnt Point) {
	drawPnt.X = float32((pnt.X-this.MapCenter.X)*this.Scale) + this.dx - this.center.X
	if this.yup {
		drawPnt.Y = float32((pnt.Y-this.MapCenter.Y)*this.Scale) + this.dy - this.center.Y
	} else {
		drawPnt.Y = this.dy - float32((pnt.Y-this.MapCenter.Y)*this.Scale) - this.center.Y
	}
	return
}

func (this *CoordParams) Forwards(pnts []base.Point2D) []Point {
	drawPnts := make([]Point, len(pnts))
	for i, v := range pnts {
		drawPnts[i] = this.Forward(v)
	}
	return drawPnts
}

// 反向转化坐标
// Reverse
