// 针对Line的算法库
package algorithm

import (
	"gogis/base"
)

// 多点折线
type Line []base.Point2D

// 抽稀一条折线
// 当两个点的距离小于阈值，且角度小于阈值时，后一个点被跳过
// 注意：dis2为距离的平方，angle为角度值（非弧度）
func (line Line) Thin(dis2 float64, angle float64) (out Line) {
	out = make([]base.Point2D, 1, len(line))
	out[0] = line[0]
	pos := 0

	count := len(line)
	for i := 1; i < count-1; i++ {
		// 距离够远，或者拐角够大的点，都应该保留下来
		if base.DistanceSquare(line[pos].X, line[pos].Y, line[i].X, line[i].Y) > dis2 ||
			(i < count-1 && base.Angle(line[pos], line[i], line[i+1]) < angle) {
			out = append(out, line[i])
			pos = i
		}
	}
	// 保留最后一个点
	out = append(out, line[count-1])
	return
}

// 瘦身，去掉同方向的中间点
func (line Line) Slim() (out Line) {
	out = make([]base.Point2D, 1, len(line))
	out[0] = line[0]
	pos := 0

	count := len(line)
	for i := 1; i < count; i++ {
		// 距离够远，或者拐角够大的点，都应该保留下来
		if i < count-1 && base.Angle(line[pos], line[i], line[i+1]) != 180 {
			out = append(out, line[i])
			pos = i
		}
	}
	// 保留最后一个点
	out = append(out, line[count-1])
	return
}

// ================================================================ //

// 判断折线是否覆盖点；即点在折线上（在端点或内部均可）
func (line Line) IsCoversPnt(pnt base.Point2D) bool {
	for i := 1; i < len(line); i++ {
		if (Segment{line[i-1], line[i]}).IsCoversPnt(pnt) {
			return true
		}
	}
	return false
}

// 判断折线是否包含点；即点在折线内部（不能在折线两头的端点上）
func (line Line) IsContainsPnt(pnt base.Point2D) bool {
	if pnt == line[0] || pnt == line[len(line)-1] {
		return false // 在端点上了
	}
	for i := 1; i < len(line); i++ {
		if (Segment{line[i-1], line[i]}).IsCoversPnt(pnt) {
			return true
		}
	}
	return false
}

// ================================================================ //

func (line Line) IsIntersectsSeg(seg Segment) bool {
	for i := 1; i < len(line); i++ {
		if (Segment{line[i-1], line[i]}).IsIntersects(seg) {
			return true
		}
	}
	return false
}

// 线覆盖线段
func (line Line) IsCoversSeg(seg Segment) bool {
	// 先去掉同方向的中间点
	newLine := line.Slim()
	// 然后seg必须在line的某个线段中才行
	for i := 1; i < len(line); i++ {
		if (Segment{newLine[i-1], newLine[i]}).IsCovers(seg) {
			return true
		}
	}
	return false
}

// 线和线段的关系，返回交集的维度，没有交集返回-1
// exFirst: 是否seg的第一个节点不参与计算
// func (line Line) CalcIIDimSeg(seg Segment, exFirst bool) (dim int) {
// 	dim = -1
// 	count := len(line)
// 	if count == 2{  // todo
// 		return
// 	}
// 	for i := 2; i < count-1; i++ {
// 		dim = base.IntMax(dim, Segment{line[i-1], line[i]}.IntersectsDim(seg, false, exFirst))
// 		if dim == 1 {
// 			return
// 		}
// 	}
// 	dim = base.IntMax(dim, Segment{line[0], line[1]}.IntersectsDim(seg, true, exFirst))
// 	if dim == 1 {
// 		return
// 	}
// 	dim = base.IntMax(dim, Segment{line[count-1], line[count-2]}.IntersectsDim(seg, true, exFirst))
// 	return
// }

// ================================================================ //

// 折线是否与rect相交
func (line Line) IsIntersectsRect(rect base.Rect2D) bool {
	for i := 1; i < len(line); i++ {
		if (Segment{line[i-1], line[i]}).IsIntersectsRect(rect) {
			return true
		}
	}
	return false
}

// 折线的内部和rect的内部有交集，还和rect的外部也有交集
func (line Line) IsCrossesRect(rect base.Rect2D) bool {
	// 首先看是否同时有节点在rect内和外
	var inRect, outRect bool
	for _, v := range line {
		if rect.IsContainsPnt(v) {
			inRect = true
		} else {
			outRect = true
		}
		// rect的内外都有就OK
		if inRect && outRect {
			return true
		}
	}
	// 如果所有点都在rect内部(故而out为false)，那肯定和rect外部无关了
	if !outRect {
		return false
	}

	// 到这里,意味着所有点都在外部,则看是否有线段cross rect了
	for i := 1; i < len(line); i++ {
		seg := Segment{line[i], line[i-1]}
		if seg.IsCrossesRect(rect) {
			return true
		}
	}

	return false
}

// 折线和rect有接触，即折线和rect有交集，且折线和rect的内部不能有交集
func (line Line) IsTouchesRect(rect base.Rect2D) (res bool) {
	// 某个线段和rect touch，且所有线段都没有穿越rect
	for i := 1; i < len(line); i++ {
		seg := Segment{line[i], line[i-1]}
		if !seg.IsCrossesRect(rect) {
			return false
		}
		if !res && seg.IsTouchesRect(rect) {
			res = true
		}
	}
	return
}

// ================================================================ //

// 两个折线是否相等；内部没有做bbox预判
// 允许两个折线的点数不同，允许顺序相反
func (line Line) IsEquals(line2 Line) bool {
	// 先看两个端点是否一致（允许顺序调换）
	if !(Segment{line[0], line[len(line)-1]}).IsEquals(Segment{line2[0], line2[len(line2)-1]}) {
		return false
	}
	return Ring(line).IsEquals(Ring(line2))
}

// 折线和折线内部和内部有交叉，即“穿过去了”；折线的中间节点挨着另一条折线，又折返回去的不算
// func (line Line) IsCrosses2(line2 Line) bool {
// 	// todo
// }

// 有交集，挨着就算
func (line Line) IsIntersects(line2 Line) bool {
	// 先看 bbox
	if !base.ComputeBounds(line).IsIntersects(base.ComputeBounds(line2)) {
		return false
	}
	for i := 1; i < len(line2); i++ {
		if line.IsIntersectsSeg(Segment{line2[i], line2[i-1]}) {
			return true
		}
	}
	return false
}

// 是否覆盖
func (line Line) IsCovers(line2 Line) bool {
	for i := 1; i < len(line2); i++ {
		if !line.IsCoversSeg(Segment{line2[i], line2[i-1]}) {
			return false
		}
	}
	return true
}

// 线和线内部的关系，返回交集的维度，没有交集返回-1
func (line Line) CalcIIDim(line2 Line) (dim int) {
	dim = -1
	c1 := len(line)
	c2 := len(line2)
	for i := 1; i < c1; i++ {
		seg1 := Segment{line[i], line[i-1]}
		for j := 1; j < c2; j++ {
			seg2 := Segment{line2[j], line2[j-1]}
			if seg1.IsOverlaps(seg2) || seg1.IsEquals(seg2) {
				dim = 1
				return
			} else if dim != 0 && seg1.IsTouches(seg2) {
				// 这里还需要排除折线的两个端点
				if !((i == 1 && seg2.IsCoversPnt(line[0])) || (j == 1 && seg1.IsCoversPnt(line2[0])) ||
					(i == c1-1 && seg2.IsCoversPnt(line[c1-1])) || (j == c2-1 && seg1.IsCoversPnt(line2[c2-1]))) {
					dim = 0
				}
			}
		}
	}
	return

	// count := len(line2)
	// // 一开始排除两头的线段，因为端点属于折线的边界
	// for i := 2; i < count-1; i++ {
	// 	dim = base.IntMax(dim, line.CalcIIDimSeg(Segment{line2[i], line2[i-1]}, false))
	// 	if dim == 1 {
	// 		return
	// 	}
	// }
	// // 再看两头的线段情况
	// dim = base.IntMax(dim, line.CalcIIDimSeg(Segment{line2[0], line2[1]}, true))
	// if dim == 1 {
	// 	return
	// }
	// dim = base.IntMax(dim, line.CalcIIDimSeg(Segment{line2[count-1], line2[count-2]}, true))
	// return
}
