package gogis

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// 坐标转化参数
type CoordParams struct {
	scale      float64 // 图片距离/地图距离
	mapCenter  Point2D // 地图中心点
	drawCenter Point   // 图片中心点
	dx, dy     int     // 图片的高度
}

// 根据地图的box和要绘制的大小，来初始化参数
func (this *CoordParams) Init(bbox Rect2D, dx int, dy int) {
	scaleX := (float64)(dx) / (bbox.Max.X - bbox.Min.X)
	scaleY := (float64)(dy) / (bbox.Max.Y - bbox.Min.Y)
	this.scale = math.Min(scaleX, scaleY)

	this.mapCenter.X = (bbox.Max.X + bbox.Min.X) / 2
	this.mapCenter.Y = (bbox.Max.Y + bbox.Min.Y) / 2

	this.drawCenter.X = dx / 2
	this.drawCenter.Y = dy / 2
	this.dx = dx
	this.dy = dy
	fmt.Println("params:", this)
}

// 得到当前地图的范围
func (this *CoordParams) GetBounds() Rect2D {
	var bbox Rect2D
	bbox.Min.X = this.mapCenter.X - float64(this.drawCenter.X)/this.scale
	bbox.Min.Y = this.mapCenter.Y - float64(this.drawCenter.Y)/this.scale
	bbox.Max.X = this.mapCenter.X + float64(this.drawCenter.X)/this.scale
	bbox.Max.Y = this.mapCenter.Y + float64(this.drawCenter.Y)/this.scale
	fmt.Println("CoordParams.GetBounds():", bbox)
	return bbox
}

// 正向转化一个点坐标：从地图坐标变为图片坐标
// 绘制坐标 x = (pnt2D - mapCenter)*scale + drawCenter
// 		    y = dy - ((pnt2D - mapCenter)*scale + drawCenter)
func (this *CoordParams) Forward(pnt Point2D) Point {
	var drawPnt Point
	drawPnt.X = (int)((pnt.X-this.mapCenter.X)*this.scale) + this.drawCenter.X
	drawPnt.Y = this.dy - ((int)((pnt.Y-this.mapCenter.Y)*this.scale) + this.drawCenter.Y)
	return drawPnt
}

// 反向转化坐标
// Reverse

type Canvas struct {
	// 真正用来绘制的图片
	img    *image.NRGBA
	params CoordParams
	// 转化参数
	// scaleX, scaleY float64
	// xmin, ymin     float64
	// dx, dy         int
}

// 求绝对值
func abs(x int) int {
	if x >= 0 {
		return x
	}
	return -x
}

type Point image.Point

// A Point is an X, Y coordinate pair. The axes increase right and down.
// type Point struct {
// 	X, Y int
// }

type IntPolyline struct {
	numParts int
	points   [][]Point
}

func (this *Canvas) DrawPolyline(polyline *IntPolyline) {
	for i := 0; i < polyline.numParts; i++ {
		for j := 0; j < len(polyline.points[i])-1; j++ {
			x0 := polyline.points[i][j].X
			y0 := polyline.points[i][j].Y
			x1 := polyline.points[i][j+1].X
			y1 := polyline.points[i][j+1].Y
			DrawLine(x0, y0, x1, y1, func(x, y int) {
				this.img.Set(x, y, color.RGBA{uint8(255), 0, 0, 255})
			})
		}
	}
}

// Putpixel describes a function expected to draw a point on a bitmap at (x, y) coordinates.
type Putpixel func(x, y int)

// Bresenham's algorithm, http://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm
// https://github.com/akavel/polyclip-go/blob/9b07bdd6e0a784f7e5d9321bff03425ab3a98beb/polyutil/draw.go
// TODO: handle int overflow etc.
// func (this *DC) DrawLine(x0, y0, x1, y1 int, brush Putpixel) {
func DrawLine(x0, y0, x1, y1 int, brush Putpixel) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 >= x1 {
		sx = -1
	}
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy

	for {
		brush(x0, y0)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}
