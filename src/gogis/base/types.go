package base

import "math"

type Point2D struct {
	X float64
	Y float64
}

// 计算两点间距离
func (this *Point2D) Distance(pnt Point2D) float64 {
	return math.Sqrt(math.Pow(this.X-pnt.X, 2) + math.Pow(this.Y-pnt.Y, 2))
}

// 计算两点间距离平方
// 避免开放，提高效率
func (this *Point2D) DistanceSquare(pnt Point2D) float64 {
	return math.Pow(this.X-pnt.X, 2) + math.Pow(this.Y-pnt.Y, 2)
}

type Rect2D struct {
	Min, Max Point2D
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

// 复制自己
func (this *Rect2D) Clone() (rect *Rect2D) {
	rect = new(Rect2D)
	rect.Max = this.Max
	rect.Min = this.Min
	return
}

// 得到矩形的四个顶点，顺序是左上右下
// func (this *Rect2D) ToPoints() (pnts []Point2D) {
// 	pnts = make([]Point2D, 4)
// 	pnts[0].X = this.Min.X
// 	pnts[0].Y = this.Max.Y
// 	pnts[1] = this.Min
// 	pnts[2] = this.Max
// 	pnts[3].X = this.Max.X
// 	pnts[3].Y = this.Min.Y
// 	return
// }

// 两个box合并，取并集的box
func (this *Rect2D) Union(rect Rect2D) {
	this.Min.X = math.Min(this.Min.X, rect.Min.X)
	this.Min.Y = math.Min(this.Min.Y, rect.Min.Y)
	this.Max.X = math.Max(this.Max.X, rect.Max.X)
	this.Max.Y = math.Max(this.Max.Y, rect.Max.Y)
}

// 两个矩形有交集，交集可以是点、线、面
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

// 是否包括点，点不能在边界
func (this *Rect2D) IsContainsPoint(pnt Point2D) bool {
	if this.Max.X > pnt.X && pnt.X > this.Min.X && this.Max.Y > pnt.Y && pnt.Y > this.Min.Y {
		return true
	}
	return false
}

// 是否包括另一个矩形，边界不能有交集
// todo 暂时先用最简单的思路，即两个对角点都在其中来判断；未来考虑用更高效的算法
func (this *Rect2D) IsContains(rect Rect2D) bool {
	if this.IsContainsPoint(rect.Min) && this.IsContainsPoint(rect.Max) {
		return true
	}
	return false
}

// 是否覆盖点，点在边界也算
func (this *Rect2D) IsCoverPoint(pnt Point2D) bool {
	if IsBigEqual(this.Max.X, pnt.X) && IsBigEqual(pnt.X, this.Min.X) && IsBigEqual(this.Max.Y, pnt.Y) && IsBigEqual(pnt.Y, this.Min.Y) {
		return true
	}
	return false
}

// 是否覆盖另一个矩形，允许边界同时有交集
// todo 暂时先用最简单的思路，即两个对角点都在其中来判断；未来考虑用更高效的算法
func (this *Rect2D) IsCover(rect Rect2D) bool {
	if this.IsCoverPoint(rect.Min) && this.IsCoverPoint(rect.Max) {
		return true
	}
	return false
}

// 两个矩形是否有交叠，即交集必须有二维部分
func (this *Rect2D) IsOverlap(rect Rect2D) bool {
	// x、y两个方向，交集都必须大于0，则相交部分存在面积
	// overlapX := (right1-left1)+(right2-left2) - ( max(right1,right2) - min(left1,left2) )
	oX := (this.Max.X - this.Min.X) + (rect.Max.X - rect.Min.X) - (math.Max(this.Max.X, rect.Max.X) - math.Min(this.Min.X, rect.Min.X))
	oY := (this.Max.Y - this.Min.Y) + (rect.Max.Y - rect.Min.Y) - (math.Max(this.Max.Y, rect.Max.Y) - math.Min(this.Min.Y, rect.Min.Y))
	if oX > 0 && oY > 0 {
		return true
	}
	return false
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

// 返回中心点
func (this *Rect2D) Center() (center Point2D) {
	center.X = (this.Max.X + this.Min.X) / 2.0
	center.Y = (this.Max.Y + this.Min.Y) / 2.0
	return
}
