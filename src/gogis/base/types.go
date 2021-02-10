// 点、矩形等基础类型

package base

import (
	"encoding/binary"
	"math"
)

// 点、线段（两个点）、折线（一串点）、封闭环（闭合曲线）之间的判断和运算
// 定义：点Point(2D)，线段Segment，折线Line、封闭环Ring
// 定义：Line的中间点为vertex（顶点），两端的点为Endpoint（端点），所有的点统称node（节点）

// 两维点
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

type Point2Ds []Point2D

//Len()
func (s Point2Ds) Len() int {
	return len(s)
}

//Less():从低到高排序
func (s Point2Ds) Less(i, j int) bool {
	return s[i].X < s[j].X || s[i].Y < s[j].Y
}

//Swap()
func (s Point2Ds) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// ===================================================== //

// GetBounds 接口
type Bounds interface {
	GetBounds() Rect2D
}

// 矩形结构
type Rect2D struct {
	Min, Max Point2D
}

func NewRect2D(minx, miny, maxx, maxy float64) (rect Rect2D) {
	rect.Min.X = minx
	rect.Min.Y = miny
	rect.Max.X = maxx
	rect.Max.Y = maxy
	return
}

// 是否为合法的矩形
func (rect Rect2D) IsValid() bool {
	return rect.Max.X > rect.Min.X && rect.Max.Y > rect.Min.Y
}

func (rect Rect2D) GetBounds() Rect2D {
	return rect
}

func (rect Rect2D) ToBytes() []byte {
	bytes := make([]byte, 32)
	minx := math.Float64bits(rect.Min.X)
	miny := math.Float64bits(rect.Min.Y)
	maxx := math.Float64bits(rect.Max.X)
	maxy := math.Float64bits(rect.Max.Y)

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

// 得到矩形的顶点，顺序是左上右下
// closed:  是否要闭合，即重复第一个点
func (rect Rect2D) ToPoints(closed bool) (pnts []Point2D) {
	pnts = make([]Point2D, 4, 5)
	pnts[0].X = rect.Min.X
	pnts[0].Y = rect.Max.Y
	pnts[1] = rect.Min
	pnts[2].X = rect.Max.X
	pnts[2].Y = rect.Min.Y
	pnts[3] = rect.Max
	if closed {
		pnts = append(pnts, pnts[0])
	}
	return
}

// 左上角
func (rect Rect2D) LeftTop() (pnt Point2D) {
	pnt.X = rect.Min.X
	pnt.Y = rect.Max.Y
	return
}

// 右下角
func (rect Rect2D) RightBottom() (pnt Point2D) {
	pnt.X = rect.Max.X
	pnt.Y = rect.Min.Y
	return
}

// 计算面积
func (rect Rect2D) Area() float64 {
	return rect.Dx() * rect.Dy()
}

func (rect Rect2D) Dx() float64 {
	return rect.Max.X - rect.Min.X
}

func (rect Rect2D) Dy() float64 {
	return rect.Max.Y - rect.Min.Y
}

// 返回中心点
func (rect Rect2D) Center() (center Point2D) {
	center.X = (rect.Max.X + rect.Min.X) / 2.0
	center.Y = (rect.Max.Y + rect.Min.Y) / 2.0
	return
}

// === 注意：以下几个操作函数，都不直接改变自身，而是返回改变后的结果 === //

// 两个box合并，返回并集的box
func (rect Rect2D) Union(other Rect2D) Rect2D {
	rect.Min.X = math.Min(rect.Min.X, other.Min.X)
	rect.Min.Y = math.Min(rect.Min.Y, other.Min.Y)
	rect.Max.X = math.Max(rect.Max.X, other.Max.X)
	rect.Max.Y = math.Max(rect.Max.Y, other.Max.Y)
	return rect
}

// 边界进行扩展; d>0时扩大;  d<0时缩小
func (rect Rect2D) Extend(d float64) Rect2D {
	rect.Min.X -= d
	rect.Min.Y -= d
	rect.Max.X += d
	rect.Max.X += d
	return rect
}

// 中心点不变，进行缩放; r 为缩放比率, >1 放大, <1缩小
func (rect Rect2D) Scale(r float64) Rect2D {
	center := rect.Center()
	width := rect.Dx() / 2.0 * r
	height := rect.Dy() / 2.0 * r
	rect.Min.X = center.X - width
	rect.Min.Y = center.Y - height
	rect.Max.X = center.X + width
	rect.Max.Y = center.Y + height
	return rect
}

// === 注意：以上几个操作函数，都不直接改变自身，而是返回改变后的结果 === //

// 把一个box从中心点分为四份，返回其中一个bbox, ax/ay 为true时，坐标值增加
func (rect Rect2D) SplitByCenter(ax bool, ay bool) (newBbox Rect2D) {
	center := rect.Center()

	if ax && ay {
		newBbox.Min = center
		newBbox.Max = rect.Max
	} else if ax {
		newBbox.Min.X = center.X
		newBbox.Max.X = rect.Max.X
		newBbox.Min.Y = rect.Min.Y
		newBbox.Max.Y = center.Y
	} else if ay {
		newBbox.Min.X = rect.Min.X
		newBbox.Max.X = center.X
		newBbox.Min.Y = center.Y
		newBbox.Max.Y = rect.Max.Y
	} else {
		newBbox.Min = rect.Min
		newBbox.Max = center
	}
	return newBbox
}

// 按 x/y 切两刀, 可能得到两个小box,也有可能得到四个；都没切到时返回[]
func (rect Rect2D) SplitByXY(x float64, y float64) (bboxes []Rect2D) {
	hx, hy := false, false // 是否切中bbox
	// 竖着切中了
	if rect.Max.X > x && x > rect.Min.X {
		hx = true
	}
	// 横着切中了
	if rect.Max.Y > y && y > rect.Min.Y {
		hy = true
	}
	if hx && hy {
		bboxes = append(bboxes, splitBoxByXY(rect, x, y)...)
	} else if hx {
		bboxes = append(bboxes, splitBoxByX(rect, x)...)
	} else if hy {
		bboxes = append(bboxes, splitBoxByY(rect, y)...)
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

// 是否包括点，点不能在边界
func (rect Rect2D) IsContainsPnt(pnt Point2D) bool {
	if rect.Max.X > pnt.X && pnt.X > rect.Min.X && rect.Max.Y > pnt.Y && pnt.Y > rect.Min.Y {
		return true
	}
	return false
}

// 是否覆盖点，点在边界也算
func (rect Rect2D) IsCoversPnt(pnt Point2D) bool {
	if IsBigEqual(rect.Max.X, pnt.X) && IsBigEqual(pnt.X, rect.Min.X) &&
		IsBigEqual(rect.Max.Y, pnt.Y) && IsBigEqual(pnt.Y, rect.Min.Y) {
		return true
	}
	return false
}

// 矩形与点相交，其实就是cover
func (rect Rect2D) IsIntersectsPnt(pnt Point2D) bool {
	return rect.IsCoversPnt(pnt)
}

// 矩形与点分离，其实就是not cover
func (rect Rect2D) IsDisjointPnt(pnt Point2D) bool {
	return !rect.IsCoversPnt(pnt)
}

// 点是否正好在矩形的边界上，即点touch矩形
func (rect Rect2D) IsTouchesPnt(pnt Point2D) bool {
	// pnt或者在竖线，或者在横线上
	if ((pnt.X == rect.Min.X || pnt.X == rect.Max.X) && (pnt.Y <= rect.Max.Y && pnt.Y >= rect.Min.Y)) ||
		((pnt.Y == rect.Min.Y || pnt.Y == rect.Max.Y) && (pnt.X <= rect.Max.X && pnt.X >= rect.Min.X)) {
		return true
	}
	return false
}

// ================================================================ //

// 是否包括另一个矩形，边界不能有交集
// 暂时先用最简单的思路，即两个对角点都在其中来判断；未来考虑用更高效的算法
func (rect Rect2D) IsContains(rect2 Rect2D) bool {
	if rect.IsContainsPnt(rect2.Min) && rect.IsContainsPnt(rect2.Max) {
		return true
	}
	return false
}

// 是否覆盖另一个矩形，允许边界同时有交集
// 暂时先用最简单的思路，即两个对角点都在其中来判断；未来考虑用更高效的算法
func (rect Rect2D) IsCovers(rect2 Rect2D) bool {
	if rect.IsCoversPnt(rect2.Min) && rect.IsCoversPnt(rect2.Max) {
		return true
	}
	return false
}

// 两个矩形是否有交集，交集可以是点、线、面
func (rect Rect2D) IsIntersects(rect2 Rect2D) bool {
	zx := math.Abs(rect.Min.X + rect.Max.X - rect2.Min.X - rect2.Max.X)
	x := math.Abs(rect.Min.X-rect.Max.X) + math.Abs(rect2.Min.X-rect2.Max.X)
	zy := math.Abs(rect.Min.Y + rect.Max.Y - rect2.Min.Y - rect2.Max.Y)
	y := math.Abs(rect.Min.Y-rect.Max.Y) + math.Abs(rect2.Min.Y-rect2.Max.Y)
	if zx <= x && zy <= y {
		return true
	} else {
		return false
	}
}

// 两个矩形是否有交叠，即交集必须有二维部分
func (rect Rect2D) IsOverlaps(rect2 Rect2D) bool {
	// x、y两个方向，交集都必须大于0，则相交部分存在面积
	// overlapX := (right1-left1)+(right2-left2) - ( max(right1,right2) - min(left1,left2) )
	oX := (rect.Max.X - rect.Min.X) + (rect2.Max.X - rect2.Min.X) - (math.Max(rect.Max.X, rect2.Max.X) - math.Min(rect.Min.X, rect2.Min.X))
	oY := (rect.Max.Y - rect.Min.Y) + (rect2.Max.Y - rect2.Min.Y) - (math.Max(rect.Max.Y, rect2.Max.Y) - math.Min(rect.Min.Y, rect2.Min.Y))
	if oX > 0 && oY > 0 {
		return true
	}
	return false
}

func (rect Rect2D) IsDisjoint(rect2 Rect2D) bool {
	return !rect.IsIntersects(rect2)
}

// 包括两种情况：1）交集为点；2）交集为线
func (rect Rect2D) IsTouches(rect2 Rect2D) bool {
	// 首先交集不能是二维的, 然后还必须有交集
	if !rect.IsOverlaps(rect2) && rect.IsIntersects(rect2) {
		return true
	}
	return false
}

// 相等
func (rect Rect2D) IsEquals(rect2 Rect2D) bool {
	return rect == rect2
}

// ================================================================ //

// 两个矩形求交集；若不相交，返回的矩形不合法
func (rect Rect2D) Intersects(rect2 Rect2D) (out Rect2D) {
	// 两个最大值的较小者，作为新的最大值
	out.Max.X = math.Min(rect.Max.X, rect2.Max.X)
	out.Max.Y = math.Min(rect.Max.Y, rect2.Max.Y)
	// 两个最小值的较大者，作为新的最小值
	out.Min.X = math.Max(rect.Min.X, rect2.Min.X)
	out.Min.Y = math.Max(rect.Min.Y, rect2.Min.Y)
	return
}

// ================================================================ //

// 计算得到点串的bounds
func ComputeBounds(points []Point2D) (bbox Rect2D) {
	bbox.Init()
	for _, pnt := range points {
		bbox = bbox.UnionPnt(pnt)
		// bbox.Min.X = math.Min(bbox.Min.X, pnt.X)
		// bbox.Min.Y = math.Min(bbox.Min.Y, pnt.Y)
		// bbox.Max.X = math.Max(bbox.Max.X, pnt.X)
		// bbox.Max.Y = math.Max(bbox.Max.Y, pnt.Y)
	}
	return
}

func (rect Rect2D) UnionPnt(pnt Point2D) (bbox Rect2D) {
	bbox.Min.X = math.Min(rect.Min.X, pnt.X)
	bbox.Min.Y = math.Min(rect.Min.Y, pnt.Y)
	bbox.Max.X = math.Max(rect.Max.X, pnt.X)
	bbox.Max.Y = math.Max(rect.Max.Y, pnt.Y)
	return
}

// 合并rect数组
func UnionBounds(bboxes []Rect2D) (bbox Rect2D) {
	bbox.Init()
	for _, v := range bboxes {
		bbox = bbox.Union(v)
	}
	return
}
