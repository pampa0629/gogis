package base

import "math"

type Point2D struct {
	X float64
	Y float64
}

type Rect2D struct {
	Min, Max Point2D
	// xmin float64
	// ymin float64
	// xmax float64
	// ymax float64
}

// 初始化，使之无效，即min为浮点数最大值；max为浮点数最小值。而非均为0
func (this *Rect2D) Init() {
	this.Min.X = math.MaxFloat64
	this.Min.Y = math.MaxFloat64
	this.Max.X = -math.MaxFloat64
	this.Max.Y = -math.MaxFloat64
}

func NewRect2D(minx, miny, maxx, maxy float64) (value Rect2D) {
	value.Min.X = minx
	value.Min.Y = miny
	value.Max.X = maxx
	value.Max.Y = maxy
	return
}

// 两个box合并，取并集的box
func (this *Rect2D) Union(rect Rect2D) {
	this.Min.X = math.Min(this.Min.X, rect.Min.X)
	this.Min.Y = math.Min(this.Min.Y, rect.Min.Y)
	this.Max.X = math.Max(this.Max.X, rect.Max.X)
	this.Max.Y = math.Max(this.Max.Y, rect.Max.Y)
}

// 两个 边框是否相交
func (this *Rect2D) IsIntersect(rect Rect2D) bool {
	zx := math.Abs(this.Min.X + this.Max.X - rect.Min.X - rect.Max.X)
	x := math.Abs(this.Min.X-this.Max.X) + math.Abs(rect.Min.X-rect.Max.X)
	zy := math.Abs(this.Min.Y + this.Max.Y - rect.Min.Y - rect.Max.Y)
	y := math.Abs(this.Min.Y-this.Max.Y) + math.Abs(rect.Min.Y-rect.Max.Y)
	if zx <= x && zy <= y {
		return true
	} else {
		return false
	}
}

// 计算得到点串的bounds
func ComputeBounds(points []Point2D) Rect2D {
	var bbox Rect2D
	bbox.Init()
	for _, pnt := range points {
		bbox.Min.X = math.Min(bbox.Min.X, pnt.X)
		bbox.Min.Y = math.Min(bbox.Min.Y, pnt.Y)
		bbox.Max.X = math.Max(bbox.Max.X, pnt.X)
		bbox.Max.Y = math.Max(bbox.Max.Y, pnt.Y)
	}
	return bbox
}

// 计算面积
func (this *Rect2D) Area() float64 {
	return this.Dx() * this.Dy()
}

func (this *Rect2D) Dx() float64 {
	return this.Max.X - this.Min.X
}

func (this *Rect2D) Dy() float64 {
	return this.Max.Y - this.Min.Y
}
