// 针对Ring的算法库
package algorithm

import (
	"gogis/base"
	"sort"
)

// 多点封闭环(不带洞)
type Ring []base.Point2D

// 抽稀一个封闭区域
func (ring Ring) Thin(dis2, angle float64) (newRing Ring) {
	newRing = Ring(Line(ring).Thin(dis2, angle))
	// 不够三个点，就不要了
	count := len(newRing)
	if count < 3 {
		return nil
	}
	// 封闭区域
	if newRing[0] != newRing[count-1] {
		newRing = append(newRing, newRing[0])
	}
	return
}

// ================================================================ //

// 环是否覆盖点；点在线上也算
func (ring Ring) IsCoversPnt(pnt base.Point2D) bool {
	return ring.isIncludePnt(pnt, true)
}

// 判断环是否包含点；即点是否在一个环之内，若点在线上，不属于在环内
// 采用射线法，看水平方向射线与所有边的交点个数的奇偶性；奇数为内，偶数为外
// func PntIsWithinRing(pnt base.Point2D, ring Ring) bool {
func (ring Ring) IsContainsPnt(pnt base.Point2D) bool {
	return ring.isIncludePnt(pnt, false)
}

// 是否包括点；cover：是否允许点在边线上
func (ring Ring) isIncludePnt(pnt base.Point2D, cover bool) bool {
	count := 0
	for i := 1; i < len(ring); i++ {
		seg := Segment{ring[i-1], ring[i]}
		// 允许在点在线上
		if cover && seg.IsCoversPnt(pnt) {
			return true
		} else if !cover && seg.IsCoversPnt(pnt) {
			return false
		}
		count += seg.CountPntRay(pnt)
	}

	if count%2 == 0 {
		return false
	}
	return true
}

// 判断环是否包括几个点
func (ring Ring) IsContainsPnts(pnts []base.Point2D) bool {
	for _, pnt := range pnts {
		if !ring.IsContainsPnt(pnt) {
			return false
		}
	}
	return true
}

// 判断环是否覆盖几个点
func (ring Ring) IsCoversPnts(pnts []base.Point2D) bool {
	for _, pnt := range pnts {
		if !ring.IsCoversPnt(pnt) {
			return false
		}
	}
	return true
}

// 判断环是否覆盖几个点的任意一个
func (ring Ring) IsCoversAnyPnts(pnts []base.Point2D) bool {
	for _, pnt := range pnts {
		if ring.IsCoversPnt(pnt) {
			return true
		}
	}
	return false
}

// ================================================================ //
// 环是否覆盖线段
func (ring Ring) IsCoversSeg(seg Segment) bool {
	if !ring.IsCoversPnt(seg.P1) || !ring.IsCoversPnt(seg.P2) {
		return false
	}
	// 	点集pointSet初始化为空
	onPnts := make([]base.Point2D, 0)
	for i := 1; i < len(ring); i++ {
		seg2 := Segment{ring[i], ring[i-1]}
		if seg2.IsCrosses(seg) {
			return false
		} else if seg2.IsCoversPnt(seg.P1) {
			onPnts = append(onPnts, seg.P1)
		} else if seg2.IsCoversPnt(seg.P2) {
			onPnts = append(onPnts, seg.P2)
		} else if seg.IsCoversPnt(seg2.P1) {
			onPnts = append(onPnts, seg2.P1)
		} else if seg.IsCoversPnt(seg2.P2) {
			onPnts = append(onPnts, seg2.P2)
		}
	}
	// 将pointSet中的点按照X-Y坐标排序；
	sort.Sort(base.Point2Ds(onPnts))
	for i := 1; i < len(onPnts); i++ {
		var cPnt base.Point2D
		cPnt.X = (onPnts[i-1].X + onPnts[i].X) / 2.0
		cPnt.Y = (onPnts[i-1].Y + onPnts[i].Y) / 2.0
		// 这里关键是要看看端点连线的中间点，是否也在面内
		if !ring.IsCoversPnt(cPnt) {
			return false
		}
	}
	return true
}

// ================================================================ //

// 环的边界与线内部的关系，返回交集的维度，没有交集返回-1
func (ring Ring) CalcBIDim(line Line) (dim int) {
	dim = -1
	c1 := len(ring)
	c2 := len(line)
	for i := 1; i < c1; i++ {
		seg1 := Segment{ring[i], ring[i-1]}
		for j := 1; j < c2; j++ {
			seg2 := Segment{line[j], line[j-1]}
			if seg1.IsOverlaps(seg2) || seg1.IsEquals(seg2) {
				dim = 1
				return
			} else if dim != 0 && seg1.IsTouches(seg2) {
				// 这里还需要排除折线的两个端点
				if !((j == 1 && seg1.IsCoversPnt(line[0])) || (j == c2-1 && seg1.IsCoversPnt(line[c2-1]))) {
					dim = 0
				}
			}
		}
	}
	return
	// count := len(line)
	// // 一开始排除两头的线段，因为端点属于折线的边界
	// for i := 2; i < count-1; i++ {
	// 	d := Line(ring).CalcIIDimSeg(Segment{line[i], line[i-1]}, false)
	// 	dim = base.IntMax(dim, d)
	// 	if dim == 1 {
	// 		return
	// 	}
	// }
	// // 再看两头的线段情况
	// dim = base.IntMax(dim, Line(ring).CalcIIDimSeg(Segment{line[0], line[1]}, true))
	// if dim == 1 {
	// 	return
	// }
	// dim = base.IntMax(dim, Line(ring).CalcIIDimSeg(Segment{line[count-1], line[count-2]}, true))
	// return
}

// ================================================================ //

// 环是否和rect touch
func (ring Ring) IsTouchesRect(rect base.Rect2D) bool {
	// rect覆盖了ring的范围，肯定不行
	if rect.IsCovers(base.ComputeBounds(ring)) {
		return false
	}
	// 边界穿越了，肯定也不行
	if Line(ring).IsCrossesRect(rect) {
		return false
	}
	// rect的角点，有一个在ring的边界上，即为touch
	cPnts := rect.ToPoints(false)
	for _, cPnt := range cPnts {
		if Line(ring).IsCoversPnt(cPnt) {
			return true
		}
	}
	// 或者ring上的点，有一个在rect边界上，也是touch
	for _, pnt := range ring {
		if rect.IsTouchesPnt(pnt) {
			return true
		}
	}
	return false
}

// 环覆盖rect，允许rect在环内的边界上
func (ring Ring) IsCoversRect(rect base.Rect2D) bool {
	if !base.ComputeBounds(ring).IsCovers(rect) {
		return false
	}
	// 角点都必须在ring之内
	if !ring.IsCoversPnts(rect.ToPoints(false)) {
		return false
	}
	// 另外，ring的边界还不能穿越rect
	if Line(ring).IsCrossesRect(rect) {
		return false
	}
	return true
}

// ================================================================ //

// 两个环是否相等；内部没有做bbox预判
// 允许两个环的点数不同，允许顺序相反
func (ring Ring) IsEquals(ring2 Ring) bool {
	// 要求某个环的所有点，都在另一个环上；反之亦然
	for i := 1; i < len(ring)-1; i++ {
		if !Line(ring2).IsCoversPnt(ring[i]) {
			return false
		}
	}
	for i := 1; i < len(ring2)-1; i++ {
		if !Line(ring).IsCoversPnt(ring2[i]) {
			return false
		}
	}
	return true
}

// 环覆盖环
func (ring Ring) IsCover(ring2 Ring) bool {
	// 先用范围预判
	if !base.ComputeBounds(ring).IsCovers(base.ComputeBounds(ring2)) {
		return false
	}
	// 每个点都在环内
	for _, v := range ring2 {
		if !ring.IsCoversPnt(v) {
			return false
		}
	}
	// 每个线段都在环内
	for i := 1; i < len(ring2); i++ {
		if !ring.IsCoversSeg(Segment{ring2[i], ring2[i-1]}) {
			return false
		}
	}
	return true
}

// 两个环的内部是否有交集
func (ring Ring) IsII(ring2 Ring) bool {
	// 相互之间有一个节点在对方内部，就是相交
	for _, v := range ring {
		if ring2.IsContainsPnt(v) {
			return true
		}
	}
	for _, v := range ring2 {
		if ring.IsContainsPnt(v) {
			return true
		}
	}

	// 各自的线段，只要有一个内部相交，就OK
	for i := 1; i < len(ring); i++ {
		for j := 1; j < len(ring2); j++ {
			seg1 := Segment{ring[i], ring[i-1]}
			seg2 := Segment{ring2[i], ring2[i-1]}
			if seg1.IsCrosses(seg2) {
				return true
			}
		}
	}

	// 还有就是包含对方的一个线段，也OK
	for i := 1; i < len(ring2); i++ {
		if ring.IsCoversSeg(Segment{ring2[i], ring2[i-1]}) {
			return true
		}
	}
	for i := 1; i < len(ring); i++ {
		if ring2.IsCoversSeg(Segment{ring[i], ring[i-1]}) {
			return true
		}
	}
	return false
}

// 两个环的边界的关系，返回交集的维度，没有交集返回-1
func (ring Ring) CalcBBDim(ring2 Ring) (dim int) {
	dim = -1
	c1 := len(ring)
	c2 := len(ring2)
	for i := 1; i < c1; i++ {
		seg1 := Segment{ring[i], ring[i-1]}
		for j := 1; j < c2; j++ {
			seg2 := Segment{ring2[j], ring2[j-1]}
			if seg1.IsOverlaps(seg2) || seg1.IsEquals(seg2) {
				dim = 1
				return
			} else if dim != 0 && seg1.IsTouches(seg2) {
				dim = 0
			}
		}
	}
	return
	// dim = -1
	// count := len(ring)
	// for i := 1; i < count; i++ {
	// 	d := Line(ring).CalcIIDimSeg(Segment{ring2[i], ring2[i-1]}, false)
	// 	dim = base.IntMax(dim, d)
	// 	if dim == 1 {
	// 		return
	// 	}
	// }
	// return
}
