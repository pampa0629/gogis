// 点、矩形等基础类型

package base

import (
	"encoding/binary"
	"math"
)

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

// 矩形结构
type Rect2D struct {
	Min, Max Point2D
}

func NewRect2D(minx, miny, maxx, maxy float64) (value Rect2D) {
	value.Min.X = minx
	value.Min.Y = miny
	value.Max.X = maxx
	value.Max.Y = maxy
	return
}

func (this *Rect2D) ToBytes() []byte {
	bytes := make([]byte, 32)
	minx := math.Float64bits(this.Min.X)
	miny := math.Float64bits(this.Min.Y)
	maxx := math.Float64bits(this.Max.X)
	maxy := math.Float64bits(this.Max.Y)

	binary.LittleEndian.PutUint64(bytes, minx)
	binary.LittleEndian.PutUint64(bytes[8:], miny)
	binary.LittleEndian.PutUint64(bytes[16:], maxx)
	binary.LittleEndian.PutUint64(bytes[24:], maxy)
	return bytes
}

func (this *Rect2D) FromBytes(data []byte) {
	this.Min.X = BytesToFloat64(data)
	this.Min.Y = BytesToFloat64(data[8:])
	this.Max.X = BytesToFloat64(data[16:])
	this.Max.Y = BytesToFloat64(data[24:])
}

// 初始化，使之无效，即min为浮点数最大值；max为浮点数最小值。而非均为0
func (this *Rect2D) Init() {
	this.Min.X = math.MaxFloat64
	this.Min.Y = math.MaxFloat64
	this.Max.X = -math.MaxFloat64
	this.Max.Y = -math.MaxFloat64
}

// 复制自己
func (this *Rect2D) Clone() (rect *Rect2D) {
	rect = new(Rect2D)
	rect.Max = this.Max
	rect.Min = this.Min
	return
}

// 得到矩形的四个顶点，顺序是左上右下
func (this *Rect2D) ToPoints() (pnts []Point2D) {
	pnts = make([]Point2D, 4)
	pnts[0].X = this.Min.X
	pnts[0].Y = this.Max.Y
	pnts[1] = this.Min
	pnts[2].X = this.Max.X
	pnts[2].Y = this.Min.Y
	pnts[3] = this.Max
	return
}

// 左上角
func (this *Rect2D) LeftTop() (pnt Point2D) {
	pnt.X = this.Min.X
	pnt.Y = this.Max.Y
	return
}

// 右下角
func (this *Rect2D) RightBottom() (pnt Point2D) {
	pnt.X = this.Max.X
	pnt.Y = this.Min.Y
	return
}

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

// 自身往外扩展 d
func (this *Rect2D) Extend(d float64) {
	this.Min.X -= d
	this.Min.Y -= d
	this.Max.X += d
	this.Max.X += d
}

// 把一个box从中心点分为四份，返回其中一个bbox, ax/ay 为true时，坐标值增加
func (this *Rect2D) SplitByCenter(ax bool, ay bool) (newBbox Rect2D) {
	center := this.Center()

	if ax && ay {
		newBbox.Min = center
		newBbox.Max = this.Max
	} else if ax {
		newBbox.Min.X = center.X
		newBbox.Max.X = this.Max.X
		newBbox.Min.Y = this.Min.Y
		newBbox.Max.Y = center.Y
	} else if ay {
		newBbox.Min.X = this.Min.X
		newBbox.Max.X = center.X
		newBbox.Min.Y = center.Y
		newBbox.Max.Y = this.Max.Y
	} else {
		newBbox.Min = this.Min
		newBbox.Max = center
	}
	return newBbox
}

// 按 x/y 切两刀, 可能得到两个小box,也有可能得到四个；都没切到时返回[]
func (this *Rect2D) SplitByXY(x float64, y float64) (bboxes []Rect2D) {
	hx, hy := false, false // 是否切中bbox
	// 竖着切中了
	if this.Max.X > x && x > this.Min.X {
		hx = true
	}
	// 横着切中了
	if this.Max.Y > y && y > this.Min.Y {
		hy = true
	}
	if hx && hy {
		bboxes = append(bboxes, splitBoxByXY(*this, x, y)...)
	} else if hx {
		bboxes = append(bboxes, splitBoxByX(*this, x)...)
	} else if hy {
		bboxes = append(bboxes, splitBoxByY(*this, y)...)
	}
	return
}

// 竖着切一刀，分割box，返回左右两个box
func splitBoxByX(bbox Rect2D, x float64) (bboxes []Rect2D) {
	bboxes = make([]Rect2D, 2)
	bboxes[0], bboxes[1] = bbox, bbox
	// 0: left
	bboxes[0].Max.X = x
	// 1: right
	bboxes[1].Min.X = x
	return
}

// 横着切一刀，分割box，返回上下两个box
func splitBoxByY(bbox Rect2D, y float64) (bboxes []Rect2D) {
	bboxes = make([]Rect2D, 2)
	bboxes[0], bboxes[1] = bbox, bbox
	// 0: up
	bboxes[0].Min.Y = y
	// 1: down
	bboxes[1].Max.Y = y
	return
}

// 横着，竖着都能切刀，分割box，返回上下左右四个box
func splitBoxByXY(bbox Rect2D, x float64, y float64) (bboxes []Rect2D) {
	bboxes = make([]Rect2D, 4)
	for i, _ := range bboxes {
		bboxes[i] = bbox
	}
	// 0: leftUp
	bboxes[0].Max.X = x
	bboxes[0].Min.Y = y
	// 1: leftDown
	bboxes[1].Max.X = x
	bboxes[1].Max.Y = y
	// 2: rightUp
	bboxes[2].Min.X = x
	bboxes[2].Min.Y = y
	// 3: rightDown
	bboxes[3].Min.X = x
	bboxes[3].Max.Y = y
	return
}

// ================================================================ //

// 计算得到点串的bounds
func ComputeBounds(points []Point2D) (bbox Rect2D) {
	bbox.Init()
	for _, pnt := range points {
		bbox.Min.X = math.Min(bbox.Min.X, pnt.X)
		bbox.Min.Y = math.Min(bbox.Min.Y, pnt.Y)
		bbox.Max.X = math.Max(bbox.Max.X, pnt.X)
		bbox.Max.Y = math.Max(bbox.Max.Y, pnt.Y)
	}
	return
}

// 合并rect数组
func UnionBounds(bboxes []Rect2D) (bbox Rect2D) {
	bbox.Init()
	for _, v := range bboxes {
		bbox.Union(v)
	}
	return
}
