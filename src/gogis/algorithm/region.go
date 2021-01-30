// 针对Region的算法库
package algorithm

import (
	"gogis/base"
)

// 区域(带洞不带岛) [0]为外环,后面为内环
type Region [][]base.Point2D

// 面是否覆盖点；点在边界OK
func (region Region) IsCoversPnt(pnt base.Point2D) bool {
	if !base.ComputeBounds(region[0]).IsCoversPnt(pnt) {
		return false
	}
	// 点必须被外环包含
	if !Ring(region[0]).IsCoversPnt(pnt) {
		return false
	}
	// 点在内环的内部也不行
	for i := 1; i < len(region); i++ {
		if Ring(region[i]).IsContainsPnt(pnt) {
			return false
		}
	}
	return true
}

// 面是否包含点；点在边界也不行
func (region Region) IsContainsPnt(pnt base.Point2D) bool {
	if !base.ComputeBounds(region[0]).IsContainsPnt(pnt) {
		return false
	}
	// 点必须被外环包含
	if !Ring(region[0]).IsContainsPnt(pnt) {
		return false
	}
	// 点在内环的内部或边界上，也不行
	for i := 1; i < len(region); i++ {
		if Ring(region[i]).IsCoversPnt(pnt) {
			return false
		}
	}
	return true
}

// 点在面的边界上，内环边界也可以
func (region Region) IsTouchesPnt(pnt base.Point2D) bool {
	if !base.ComputeBounds(region[0]).IsCoversPnt(pnt) {
		return false
	}
	for _, v := range region {
		if Line(v).IsCoversPnt(pnt) {
			return true
		}
	}
	return false
}

// ================================================================ //

// 面是否覆盖线段
func (region Region) IsCoversSeg(seg Segment) bool {
	if !region.IsCoversPnt(seg.P1) || !region.IsCoversPnt(seg.P2) {
		return false
	}
	if !Ring(region[0]).IsCoversSeg(seg) {
		return false
	}
	for i := 1; i < len(region); i++ {
		ring := Ring(region[i])
		if ring.IsContainsPnt(seg.P1) || ring.IsContainsPnt(seg.P2) || ring.IsCoversSeg(seg) {
			return false
		}
		for j := 1; j < len(ring); j++ {
			if (Segment{ring[j], ring[j-1]}).IsCrosses(seg) {
				return false
			}
		}
	}
	return true
}

// ================================================================ //

// 带洞的多边形是否覆盖矩形
func (region Region) IsCoversRect(rect base.Rect2D) bool {
	// 首先是bbox覆盖
	if base.ComputeBounds(region[0]).IsCovers(rect) {
		// 矩形的四个角点都必须在外环内（可在边界）
		cPnts := rect.ToPoints(false)
		if !Ring(region[0]).IsCoversPnts(cPnts) {
			return false
		}
		// 矩形的四条边不能和外环相交
		if !Line(region[0]).IsIntersectsRect(rect) {
			return false
		}

		for i := 1; i < len(region); i++ {
			ring := Ring(region[i])
			// 矩形的四个角点任一都不能在内环内（可在边界）
			for _, v := range cPnts {
				if ring.IsCoversPnt(v) {
					return false
				}
			}
			// 矩形和内环不能相互覆盖
			subBbox := base.ComputeBounds(ring)
			if subBbox.IsCovers(rect) || rect.IsCovers(subBbox) {
				return false
			}
			// 矩形和内环边界不能相交
			if Line(ring).IsIntersectsRect(rect) {
				return false
			}
		}
		return true
	}
	return false
}

// 带洞的多边形，是否和rect touch；
// 分为两种情况：1）rect和外环touch；2）rect和某一个内环touch；
func (region Region) IsTouchesRect(rect base.Rect2D) bool {
	// 如果rect 覆盖region的bbox，则必然内部有交集了
	if rect.IsCovers(base.ComputeBounds(region[0])) {
		return false
	}
	// rect的角点不能在面内
	cPnts := rect.ToPoints(false)
	for _, pnt := range cPnts {
		if region.IsContainsPnt(pnt) {
			return false
		}
	}
	// 内环、外环的线，都不能cross rect
	for _, ring := range region {
		if Line(ring).IsCrossesRect(rect) {
			return false
		}
	}

	// 外环touch rect即可
	if Ring(region[0]).IsTouchesRect(rect) {
		return true
	}
	// 内环的touch判断比较麻烦
	for i := 1; i < len(region); i++ {
		// 首先必须bbox覆盖矩形
		if base.ComputeBounds(region[i]).IsCovers(rect) {
			// 如果rect的角点中有一个在内环的边上，即为touch（前面已经排除了 rect与ring cross的情况
			for _, cPnt := range cPnts {
				if Line(region[i]).IsCoversPnt(cPnt) {
					return true
				}
			}
		}
	}
	return false
}

// 内部有部分交集
func (region Region) IsOverlapsRect(rect base.Rect2D) bool {
	bbox := base.ComputeBounds(region[0])
	// 范围必须相交
	if !bbox.IsIntersects(rect) {
		return false
	}
	// rect不能把region都覆盖了
	if rect.IsCovers(bbox) {
		return false
	}
	// 如果外环的边界穿越rect，那么肯定ok
	if Line(region[0]).IsCrossesRect(rect) {
		return true
	}
	// 剩下的情况分为两种：1）rect和region无关；2）外环覆盖rect，这时要判断rect是否有部分在内环之中
	cPnts := rect.ToPoints(false)
	// 有一个角点被外环覆盖，即可断定rect都在外环内部了
	if Ring(region[0]).IsCoversAnyPnts(cPnts) {
		// 通过是否有角点在内环内，来判断rect是否有一部分在内环内部
		cpInSubRing := false
		for _, cPnt := range cPnts {
			for i := 1; i < len(region); i++ {
				if Ring(region[i]).IsContainsPnt(cPnt) {
					cpInSubRing = true
					break
				}
			}
			if cpInSubRing {
				break
			}
		}
		// 没有角点在内环内，那肯定不行
		if !cpInSubRing {
			return false
		}
		// rect都在某单一内环所覆盖，也不行
		for i := 1; i < len(region); i++ {
			if Ring(region[i]).IsCoversRect(rect) {
				return false
			}
		}
		return true
	}
	return false
}

// ================================================================ //

// 面是否覆盖折线
func (region Region) IsCoversLine(line Line) bool {
	// 先做bbox判断
	if !base.ComputeBounds(region[0]).IsIntersects(base.ComputeBounds(line)) {
		return false
	}
	// 折线的所有节点，都在region内(可在边界)
	for _, pnt := range line {
		if !region.IsCoversPnt(pnt) {
			return false
		}
	}
	// 折线的所有线段，都必须在面内
	for i := 1; i < len(line); i++ {
		if !region.IsCoversSeg(Segment{line[i], line[i-1]}) {
			return false
		}
	}
	return true
}

// ================================================================ //

// 多边形（有洞无岛）是否一致；内部未做bbox预判
func (region Region) IsEquals(region2 Region) bool {
	// 要求子对象数量一样
	if len(region) != len(region2) {
		return false
	}
	if !Ring(region[0]).IsEquals(region2[0]) {
		return false
	}

	if len(region) > 1 {
		// 对可能的洞进行判断
		used := make([]bool, len(region))
		for i := 1; i < len(region); i++ {
			bbox1 := base.ComputeBounds(region[i])
			for j := 1; j < len(region2); j++ {
				if !used[j] {
					bbox2 := base.ComputeBounds(region2[j])
					if bbox1 == bbox2 {
						used[j] = true
						if !Ring(region[i]).IsEquals(region2[j]) {
							return false
						}
					}
				}
			}
		}
	}

	return true
}

// 面面内部有有交集
func (region Region) IsII(region2 Region) bool {
	// 先做外环范围的预判
	if !base.ComputeBounds(region[0]).IsOverlaps(base.ComputeBounds(region2[0])) {
		return false
	}
	// 如果外环的内部有交集，
	if Ring(region[0]).IsII(Ring(region2[0])) {
		// 且一个region的外环没有被另一个region的某个内环所cover，
		for i := 1; i < len(region2); i++ {
			if Ring(region2[i]).IsCover(Ring(region[0])) {
				return false
			}
		}
		for i := 1; i < len(region); i++ {
			if Ring(region[i]).IsCover(Ring(region2[0])) {
				return false
			}
		}
		// 则内部有交集
		return true
	}
	return false
}

// 面覆盖面
func (region Region) IsCovers(region2 Region) bool {
	if !base.ComputeBounds(region[0]).IsCovers(base.ComputeBounds(region2[0])) {
		return false
	}
	// a的shell必须包括b的shell
	if !Ring(region[0]).IsCover(region2[0]) {
		return false
	}

	// a的holes不能和b的shell有内部交集
	for i := 1; i < len(region); i++ {
		if Ring(region[i]).IsII(region2[0]) {
			return false
		}
	}
	return true
}
