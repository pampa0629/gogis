// 针对两点线段的算法库
package algorithm

import (
	"gogis/base"
	"math"
)

// 两点线段
type Segment struct {
	P1, P2 base.Point2D
}

// 判断线段是否覆盖点；包括点和线段的端点重合
// 细分为两种情况：点在线段内 [T********] 和 点在线段边界 [*T*******]
func (seg Segment) IsCoversPnt(pnt base.Point2D) bool {
	// 等于某个节点，自然在了
	if pnt == seg.P1 || pnt == seg.P2 {
		return true
	}
	return seg.IsContainsPnt(pnt)
}

// 线段是否包含点；即点是否在线段上；注：点不能与线段的两个端点重合
func (seg Segment) IsContainsPnt(pnt base.Point2D) bool {
	// 这里先假定两个节点不能相同
	if seg.P1 == seg.P2 {
		panic("two point of a segment cannot be same.")
	}
	if pnt == seg.P1 || pnt == seg.P2 {
		return false
	}
	// 点必须在两个点的范围内
	if pnt.X >= math.Min(seg.P1.X, seg.P2.X) && pnt.X <= math.Max(seg.P1.X, seg.P2.X) &&
		pnt.Y >= math.Min(seg.P1.Y, seg.P2.Y) && pnt.Y <= math.Max(seg.P1.Y, seg.P2.Y) {
		if base.IsEqual(base.CrossX(seg.P1, seg.P2, pnt), 0) {
			return true
		}
	}
	return false
}

// ================================================================ //

// todo
// 线段的相交关系，返回交集的维度，没有交集返回-1
// func (seg Segment) IntersectsDim(seg2 Segment, exFirst1, exFirst2 bool) (dim int) {
// 	dim = -1
// 	if seg.IsCrosses(seg2) {
// 		dim = 0
// 	}

// 	count := calcNodesTouchCount(seg, seg2)
// 	switch count {
// 	case 1:
// 		dim = 0
// 		// 对于只有一个端点在对方线段上，还要看是否需要排除
// 		if (exFirst1 && seg2.IsCoversPnt(seg.P1)) ||
// 			(exFirst2 && seg.IsCoversPnt(seg2.P1)) {
// 			dim = -1
// 		}
// 	case 2:
// 		dim = 1
// 		// 如果正好是两个端点挨着，交集也只有这个点而已
// 		if seg.P1 == seg2.P1 || seg.P1 == seg2.P2 || seg.P2 == seg2.P1 || seg.P2 == seg2.P2 {
// 			dim = 0
// 		}
// 	case 3, 4:
// 		// 超过2个点，说明肯定有内部重叠了
// 		dim = 1
// 	}
// 	return
// }

// 线段覆盖
func (seg Segment) IsCovers(seg2 Segment) bool {
	if seg.IsCoversPnt(seg2.P1) && seg.IsCoversPnt(seg2.P2) {
		return true
	}
	return false
}

func calcNodesTouchCount(seg1, seg2 Segment) (count int) {
	if seg2.IsCoversPnt(seg1.P1) {
		count++
	}
	if seg2.IsCoversPnt(seg1.P2) {
		count++
	}
	if seg1.IsCoversPnt(seg2.P1) {
		count++
	}
	if seg1.IsCoversPnt(seg2.P2) {
		count++
	}
	return count
}

// 线段与线段是否touch: 内部不能有交集；某个端点必须在另一个线段上（含该线段的端点）
func (seg Segment) IsTouches(seg2 Segment) bool {
	// 方法：计算端点在另一个线段上的数量，该数量必须为1；若为2，则除非两个端点正好挨着
	count := calcNodesTouchCount(seg, seg2)
	if count > 2 || count == 0 {
		return false
	} else if count == 1 {
		return true
	} else if count == 2 {
		return seg.P1 == seg2.P1 || seg.P1 == seg2.P2 || seg.P2 == seg2.P1 || seg.P2 == seg2.P2
	}
	return false
}

// 两个端点是否一致（允许顺序调换）
func (seg Segment) IsEquals(seg2 Segment) bool {
	if (seg.P1 == seg2.P1 && seg.P2 == seg2.P2) || (seg.P1 == seg2.P2 && seg.P2 == seg2.P1) {
		return true
	}
	return false
}

// 两个线段是否交叉（内部的交集为点）；端点在线段上不算，线段有重叠不算
// 用叉乘法计算，具体参见：https://www.cnblogs.com/tuyang1129/p/9390376.html
func (seg Segment) IsCrosses(seg2 Segment) bool {
	if base.CrossX(seg2.P1, seg2.P2, seg.P1)*base.CrossX(seg2.P1, seg2.P2, seg.P2) >= 0 {
		return false
	}
	if base.CrossX(seg.P1, seg.P2, seg2.P1)*base.CrossX(seg.P1, seg.P2, seg2.P2) >= 0 {
		return false
	}
	return true
}

// 重叠（内部部分相交），相互不能覆盖
func (seg Segment) IsOverlaps(seg2 Segment) bool {
	on21 := seg.IsContainsPnt(seg2.P1)
	on22 := seg.IsContainsPnt(seg2.P2)
	on11 := seg2.IsContainsPnt(seg.P1)
	on12 := seg2.IsContainsPnt(seg.P2)
	if (on21 || on22) && (on11 || on12) {
		// 一个线段的两个端点，不能都在另一个线段上
		if (on21 && on22) || (on11 && on12) {
			return false
		}
		return true
	}
	return false
}

// 有交集就成
func (seg Segment) IsIntersects(seg2 Segment) bool {
	if seg.IsCoversPnt(seg2.P1) {
		return true
	}
	if seg.IsCoversPnt(seg2.P2) {
		return true
	}
	if seg2.IsCoversPnt(seg.P1) {
		return true
	}
	if seg2.IsCoversPnt(seg.P2) {
		return true
	}
	return seg.IsCrosses(seg2)
}

// 线段叉乘
// func (seg Segment) CrossX(seg2 Segment) float64 {
// 	x1 := seg.P1.X - seg.P2.X
// 	y1 := seg.P1.Y - seg.P2.Y
// 	x2 := seg2.P1.X - seg2.P2.X
// 	y2 := seg2.P1.Y - seg2.P2.Y
// 	return x1*y2 - y1*x2
// }
// double compute(double x1,double y1,double x2,double y2)
//     return x1*y2 - y1*x2;

// ================================================================ //

// 线段是否与rect相交； 端点和rect挨着，也算相交
// 先看线段所在直线是否与矩形相交，
// * 如果不相交则返回false，
// * 如果相交，
//   * 则看线段的两个点是否在矩形的同一边（即两点的x(y)坐标都比矩形的小x(y)坐标小，或者大）,
//     * 若在同一边则返回false，
//     * 否则就是相交的情况。
func (seg Segment) IsIntersectsRect(rect base.Rect2D) bool {
	if !SLineIsIntersectsBbox(seg.P1, seg.P2, rect) {
		return false
	}
	// 如果线段的两个点，在矩形同一侧，则不相交
	if (seg.P1.X < rect.Min.X && seg.P2.X < rect.Min.X) ||
		(seg.P1.X > rect.Max.X && seg.P2.X > rect.Max.X) ||
		(seg.P1.Y < rect.Min.Y && seg.P2.Y < rect.Min.Y) ||
		(seg.P1.Y > rect.Max.Y && seg.P2.Y > rect.Max.Y) {
		return false
	}
	return true
}

// 穿越;分为:1)一个点在内,一个点在外; 2)线段与rect的任一边内部相交(端点在线上不算)
func (seg Segment) IsCrossesRect(rect base.Rect2D) bool {
	c1 := rect.IsContainsPnt(seg.P1)
	c2 := rect.IsContainsPnt(seg.P1)
	if (c1 && !c2) || (c2 && !c1) {
		return true
	}

	// 方法1：判断线段与rect的四条边是否相交，有一个就ok
	cPnts := rect.ToPoints(true)
	for i := 1; i < len(cPnts); i++ {
		if seg.IsCrosses(Segment{cPnts[i], cPnts[i-1]}) {
			return true
		}
	}
	return false

	// 方法2：若线段的端点在rect边界上，或者rect角点在线段上，返回false
	// cPnts := rect.ToPoints(true)
	// for i := 1; i < len(cPnts); i++ {
	// 	rectSeg := Segment{cPnts[i], cPnts[i-1]}
	// 	if rectSeg.IsCoversPnt(seg.P1) || rectSeg.IsCoversPnt(seg.P2) {
	// 		return false
	// 	}
	// }
	// for _, v := range cPnts {
	// 	if seg.IsCoversPnt(v) {
	// 		return false
	// 	}
	// }
	// // 然后线段还必须与rect相交
	// if seg.IsIntersectsRect(rect) {
	// 	return true
	// }
}

// 内部不能有交集,内部与边界、边界与边界有交集
func (seg Segment) IsTouchesRect(rect base.Rect2D) bool {
	if rect.IsContainsPnt(seg.P1) {
		return false
	}
	if rect.IsContainsPnt(seg.P2) {
		return false
	}
	// 如果线段和rect边界相交，返回false
	cPnts := rect.ToPoints(true)
	for i := 1; i < len(cPnts); i++ {
		if seg.IsCrosses(Segment{cPnts[i], cPnts[i-1]}) {
			return false
		}
	}
	// 然后，还至少有线段的端点在rect边界上，或者rect角点在线段上
	for i := 1; i < len(cPnts); i++ {
		rectSeg := Segment{cPnts[i], cPnts[i-1]}
		if rectSeg.IsCoversPnt(seg.P1) || rectSeg.IsCoversPnt(seg.P2) {
			return true
		}
	}
	for _, v := range cPnts {
		if seg.IsCoversPnt(v) {
			return true
		}
	}
	return false
}

// ================================================================ //

// 点往X+做射线，看是否与线段有交点；有返回1，没有返回0
// 该函数一般用于判断点是否在面内（射线法）
// 如果射线穿过Y值较小的端点，返回1；否则返回0 （避免判断点在面内时，这个端点重复计算）
func (seg Segment) CountPntRay(pnt base.Point2D) int {
	// 水平的线段，认为没有交点（交点可能有无限多个）
	if seg.P1.Y == seg.P2.Y {
		return 0
	}
	// 如果点的射线，与一个节点相交，则忽略y值较小的情况
	miny := math.Min(seg.P1.Y, seg.P2.Y)
	maxy := math.Max(seg.P1.Y, seg.P2.Y)
	if pnt.Y == seg.P1.Y || pnt.Y == seg.P2.Y {
		if pnt.Y == miny {
			return 0
		} else {
			return 1
		}
	}
	// 如果点的Y 在 p1.y 和 p2.y 之间，
	if pnt.Y > miny && pnt.Y < maxy {
		// 且点的X 比 点水平线与线段交点的X小，则说明有交点
		xinters := (pnt.Y-seg.P1.Y)*(seg.P2.X-seg.P1.X)/(seg.P2.Y-seg.P1.Y) + seg.P1.X
		if pnt.X < xinters {
			return 1
		}
	}
	return 0
}

// 判断直线(straight line)是否与线段相交
// 用叉乘法：若直线和线段相交，则线段两个端点分别与直线的两侧，即叉乘之积为负数
// 等于0，则直线与线段在端点相交，返回true
func (seg Segment) IsIntersectsSLine(s1, s2 base.Point2D) bool {
	x1 := base.CrossX(s1, s2, seg.P1)
	x2 := base.CrossX(s1, s2, seg.P2)
	if x1*x2 <= 0 {
		return true
	}
	return false
}

// 直线与rect是否相交，有一个相交，即返回true
func SLineIsIntersectsBbox(s1, s2 base.Point2D, rect base.Rect2D) bool {
	pnts := rect.ToPoints(true)
	for i := 1; i < len(pnts); i++ {
		if (Segment{pnts[i-1], pnts[i]}).IsIntersectsSLine(s1, s2) {
			return true
		}
	}
	return false
}
