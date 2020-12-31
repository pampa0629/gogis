package draw

import (
	"gogis/base"
	"math"
)

// 坐标转化参数
type CoordParams struct {
	Scale      float64      // 图片距离/地图距离
	MapCenter  base.Point2D // 地图中心点
	drawCenter Point        // 图片中心点
	dx, dy     int          // 图片的宽度和高度
}

// 根据地图的box和要绘制的大小，来初始化参数
func (this *CoordParams) Init(bbox base.Rect2D, dx int, dy int) {
	scaleX := (float64)(dx) / (bbox.Max.X - bbox.Min.X)
	scaleY := (float64)(dy) / (bbox.Max.Y - bbox.Min.Y)
	this.Scale = math.Min(scaleX, scaleY)

	this.MapCenter.X = (bbox.Max.X + bbox.Min.X) / 2
	this.MapCenter.Y = (bbox.Max.Y + bbox.Min.Y) / 2

	this.drawCenter.X = dx / 2
	this.drawCenter.Y = dy / 2
	this.dx = dx
	this.dy = dy
	// fmt.Println("params:", this)
}

// 得到当前地图的范围
func (this *CoordParams) GetBounds() base.Rect2D {
	var bbox base.Rect2D
	bbox.Min.X = this.MapCenter.X - float64(this.drawCenter.X)/this.Scale
	bbox.Min.Y = this.MapCenter.Y - float64(this.drawCenter.Y)/this.Scale
	bbox.Max.X = this.MapCenter.X + float64(this.drawCenter.X)/this.Scale
	bbox.Max.Y = this.MapCenter.Y + float64(this.drawCenter.Y)/this.Scale
	// fmt.Println("CoordParams.GetBounds():", bbox)
	return bbox
}

// 正向转化一个点坐标：从地图坐标变为图片坐标
// 绘制坐标 x = (pnt2D - mapCenter)*scale + drawCenter
// 		    y = dy - ((pnt2D - mapCenter)*scale + drawCenter)
func (this *CoordParams) Forward(pnt base.Point2D) Point {
	var drawPnt Point
	drawPnt.X = (int)((pnt.X-this.MapCenter.X)*this.Scale) + this.drawCenter.X
	drawPnt.Y = this.dy - ((int)((pnt.Y-this.MapCenter.Y)*this.Scale) + this.drawCenter.Y)
	return drawPnt
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
