package geometry

import (
	"gogis/algorithm"
	"gogis/base"
	"strconv"
)

type GeoRelation struct {
	A     Geometry
	B     Geometry
	reqIM *base.D9IM
}

func (this *GeoRelation) IsMatch(mode base.SpatialMode) bool {
	switch mode {
	case base.BBoxIntersects:
		return this.IsBBoxIntersects()
	case base.Intersects, base.Undefined:
		return this.IsIntersects()
	case base.Disjoint:
		return this.IsDisjoint()
	case base.Equals:
		return this.IsEquals()
	case base.Overlaps:
		return this.IsOverlaps()
	case base.Touches:
		return this.IsTouches()
	case base.Crosses:
		return this.IsCrosses()
	case base.Contains:
		return this.IsContains()
	case base.Covers:
		return this.IsCovers()
	case base.Within:
		return this.IsWithin()
	case base.CoveredBy:
		return this.IsCoveredBy()
	default:
		this.Perpare(string(mode))
		im := this.CalcRelateIM()
		return im.MatchIM(*this.reqIM)
	}
	// return false
}

func (this *GeoRelation) IsEquals() bool {
	if this.A.Dim() == this.B.Dim() && this.A.DimB() == this.B.DimB() {
		this.Perpare("TFFFTFFF2")
		im := this.CalcRelateIM()
		return im.MatchIM(*this.reqIM)
	}
	return false
}

func (this *GeoRelation) IsIntersects() bool {
	this.Perpare("TTTT*****")
	im := this.CalcRelateIM()
	strs := []string{"T********", "*T*******", "***T*****", "****T****"}
	for _, str := range strs {
		if im.Match(this.Perpare(str)) {
			return true
		}
	}
	return false
}

func (this *GeoRelation) IsDisjoint() bool {
	this.Perpare("FF*FF****")
	im := this.CalcRelateIM()
	return im.MatchIM(*this.reqIM)
}

func (this *GeoRelation) IsTouches() bool {
	this.Perpare("FT*TT****")
	im := this.CalcRelateIM()
	strs := []string{"FT*******", "F**T*****", "F***T****"}
	for _, str := range strs {
		if im.Match(this.Perpare(str)) {
			return true
		}
	}
	return false
}

func (this *GeoRelation) IsCrosses() bool {
	// 线线特殊一点
	if this.A.Dim() == 1 && this.B.Dim() == 1 {
		this.Perpare("0********")
	} else {
		this.Perpare("T*T******")
	}
	im := this.CalcRelateIM()
	return im.MatchIM(*this.reqIM)
}

// 要求边界不能挨着
func (this *GeoRelation) IsContains() bool {
	this.Perpare("TTTFFTFFT")
	im := this.CalcRelateIM()
	return im.MatchIM(*this.reqIM)
}

// 边界不能挨着
func (this *GeoRelation) IsWithin() bool {
	this.Perpare("TFFTFFTTT")
	im := this.CalcRelateIM()
	return im.MatchIM(*this.reqIM)
}

// 边界可以挨着
func (this *GeoRelation) IsCoveredBy() bool {
	this.Perpare("T*FT*F**T")
	im := this.CalcRelateIM()
	return im.MatchIM(*this.reqIM)
}

// 边界可以挨着
func (this *GeoRelation) IsCovers() bool {
	this.Perpare("TT****FFT")
	im := this.CalcRelateIM()
	return im.MatchIM(*this.reqIM)
}

func (this *GeoRelation) IsOverlaps() bool {
	// 线线特殊一点
	if this.A.Dim() == 1 && this.B.Dim() == 1 {
		this.Perpare("1*T***T**")
	} else {
		this.Perpare("T*T***T**")
	}
	im := this.CalcRelateIM()

	return im.MatchIM(*this.reqIM)
}

func (this *GeoRelation) IsBBoxIntersects() bool {
	return this.A.GetBounds().IsIntersects(this.B.GetBounds())
}

// ================================================================ //

// 预处理一下，主要是看看a和b的边界维度是否为-1
func (this *GeoRelation) Perpare(str string) string {
	// var im base.D9IM
	this.reqIM = new(base.D9IM)
	this.reqIM.Init(str)
	if this.A.DimB() == -1 {
		this.reqIM.Set(base.B, base.I, '*')
		this.reqIM.Set(base.B, base.B, '*')
		this.reqIM.Set(base.B, base.E, '*')
	}
	if this.B.DimB() == -1 {
		this.reqIM.Set(base.I, base.B, '*')
		this.reqIM.Set(base.B, base.B, '*')
		this.reqIM.Set(base.E, base.B, '*')
	}
	return this.reqIM.String()
}

func (this *GeoRelation) needII() bool {
	return this.reqIM != nil && this.reqIM.Get(base.I, base.I) != '*'
}
func (this *GeoRelation) needIB() bool {
	return this.reqIM != nil && this.reqIM.Get(base.I, base.B) != '*'
}
func (this *GeoRelation) needIE() bool {
	return this.reqIM != nil && this.reqIM.Get(base.I, base.E) != '*'
}

func (this *GeoRelation) needBI() bool {
	return this.reqIM != nil && this.reqIM.Get(base.B, base.I) != '*'
}
func (this *GeoRelation) needBB() bool {
	return this.reqIM != nil && this.reqIM.Get(base.B, base.B) != '*'
}
func (this *GeoRelation) needBE() bool {
	return this.reqIM != nil && this.reqIM.Get(base.B, base.E) != '*'
}

func (this *GeoRelation) needEI() bool {
	return this.reqIM != nil && this.reqIM.Get(base.E, base.I) != '*'
}
func (this *GeoRelation) needEB() bool {
	return this.reqIM != nil && this.reqIM.Get(base.E, base.B) != '*'
}

// func (this *GeoRelation) needEI() bool {
// 	return this.reqIM != nil && this.reqIM.Get(base.E, base.I) != '*'
// }

// ================================================================ //

func (this *GeoRelation) CalcRelateIM() (im base.D9IM) {
	im.Init("FFFFFFFFF")
	im.Set(base.E, base.E, '2') // 外部和外部永远是2

	// 如果bbox都分离，那就直接得到结果了
	if this.A.GetBounds().IsDisjoint(this.B.GetBounds()) {
		this.setDisjoint(&im)
		return
	}

	switch this.A.Dim() {
	case 0:
		// 点没有边界，故而 BX 肯定为*
		im.Set(base.B, base.I, '*')
		im.Set(base.B, base.B, '*')
		im.Set(base.B, base.E, '*')
		switch this.B.Dim() {
		case 0:
			this.calc00(&im)
		case 1:
			this.calc01(&im)
		case 2:
			this.calc02(&im)
		}
	case 1:
		switch this.B.Dim() {
		case 0: // 线和点 == 点和线，再倒转矩阵
			this.calc01(&im)
			im.Invert()
		case 1:
			this.calc11(&im)
		case 2:
			this.calc12(&im)
		}
	case 2:
		switch this.B.Dim() {
		case 0:
			this.calc02(&im)
			im.Invert()
		case 1:
			this.calc12(&im)
			im.Invert()
		case 2:
			this.calc22(&im)
		}
	}
	return
}

// 面和面
func (this *GeoRelation) calc22(im *base.D9IM) {
	// 面和面的关系可以包括：1）分离；2）边界一个点/一条线；3）内部相交；4）（被）覆盖
	a := this.A.(Geo2)
	b := this.B.(Geo2)
	ca := this.A.SubCount()
	cb := this.B.SubCount()
	// 边界
	if this.needBB() {
		dim := regionsBBRegions(a, ca, b, cb)
		im.Set(base.B, base.B, dim2im(dim))
	}

	// A 覆盖 B
	acoverb := regionsAreCoverRegions(a, ca, b, cb)
	if acoverb {
		im.Set(base.I, base.I, '2')
		im.Set(base.I, base.B, '1')
		im.Set(base.I, base.E, '2')
		im.Set(base.B, base.E, '1')
	} else {
		im.Set(base.E, base.I, '2')
		im.Set(base.E, base.B, '1')
	}
	// B 覆盖 A
	bcovera := regionsAreCoverRegions(b, cb, a, ca)
	if bcovera {
		im.Set(base.I, base.I, '2')
		im.Set(base.B, base.I, '1')
		im.Set(base.E, base.I, '2')
		im.Set(base.E, base.B, '1')
	} else {
		im.Set(base.I, base.E, '2')
		im.Set(base.B, base.E, '1')
	}
	//  这里看看两个是否完全相等
	if acoverb && bcovera {
		// im.Set(base.I, base.I, '2')
		im.Set(base.B, base.I, 'F')
		im.Set(base.E, base.I, 'F')
		im.Set(base.E, base.B, 'F')
		im.Set(base.I, base.B, 'F')
		im.Set(base.I, base.E, 'F')
		im.Set(base.B, base.E, 'F')
		// if dim != 1 {
		// 	panic("dim error")
		// }
		return
	} else if !acoverb && !bcovera {
		// 看看是否内部有交集
		if regionsAreOverlapRegions(a, ca, b, cb) {
			im.Set(base.I, base.I, '2')
			im.Set(base.I, base.B, '1')
			im.Set(base.I, base.E, '2')
			im.Set(base.B, base.E, '1')
			im.Set(base.B, base.I, '1')
			im.Set(base.E, base.I, '2')
			im.Set(base.E, base.B, '1')
		}
	}
	return
}

// 面面有交集
func regionsAreOverlapRegions(a Geo2, ca int, b Geo2, cb int) bool {
	for i := 0; i < ca; i++ {
		region1 := a.GetSubRegion(i)
		for j := 0; j < cb; j++ {
			region2 := b.GetSubRegion(j)
			if region1.IsII(region2) {
				return true
			}
		}
	}
	return false
}

// 面是否覆盖面
func regionsAreCoverRegions(a Geo2, ca int, b Geo2, cb int) bool {
	for i := 0; i < cb; i++ {
		region2 := b.GetSubRegion(i)
		cover := false
		for j := 0; j < ca; j++ {
			region1 := a.GetSubRegion(j)
			if region1.IsCovers(region2) {
				cover = true
				break
			}
		}
		if !cover {
			return false
		}
	}
	return true
}

// 面的边界之间交集的维度
func regionsBBRegions(a Geo2, ca int, b Geo2, cb int) (dim int) {
	dim = -1
	for i := 0; i < ca; i++ {
		region1 := a.GetSubRegion(i)
		for _, v := range region1 {
			for j := 0; j < cb; j++ {
				region2 := b.GetSubRegion(j)
				for _, vv := range region2 {
					dim = base.IntMax(dim, algorithm.Ring(v).CalcBBDim(vv))
					if dim == 1 {
						return
					}
				}
			}
		}
	}
	return
}

// 线和面
func (this *GeoRelation) calc12(im *base.D9IM) {
	// 线的外部和面的内部肯定会有交集
	im.Set(base.E, base.B, '2')

	a := this.A.(Geo1)
	b := this.B.(Geo2)
	pntsA := a.GetPntsB()
	countA := this.A.SubCount()
	countB := this.B.SubCount()
	if this.needBI() || this.needII() || this.needBB() || this.needBE() {
		// 先看线的边界（端点）在面的内部、边界或外部的情况
		in, on, out := pntsWithRegions(pntsA, b, countB)
		if in {
			im.Set(base.B, base.I, '0')
			// 线的端点若在面内，那么线肯定也在面内了
			im.Set(base.I, base.I, '1')
		}
		if on {
			im.Set(base.B, base.B, '0')
		}
		if out {
			im.Set(base.B, base.E, '0')
		}
	}

	if this.needIE() || this.needEI() || this.needEB() || this.needIE() {
		// 判断面是否覆盖线
		if regionsIsCoversLines(b, countB, a, countA) {
			// 面若覆盖线，则线内和面外无交集
			im.Set(base.I, base.E, 'F')
		} else {
			// 面没有覆盖线，则线的外部肯定和面的内部/边界都有交集，线内和面外有交集
			im.Set(base.E, base.I, '2')
			im.Set(base.E, base.B, '1')
			im.Set(base.I, base.E, '1')
		}
	}

	if this.needIB() {
		// 线的内部，与面的边界是否有交集，维度是0 or 1
		dim := linesIBRegions(a, countA, b, countB)
		im.Set(base.I, base.B, dim2im(dim))
	}

	if this.needEB() {
		// 线的外部，是否与面的边界有交集（除非线能完全覆盖面的边界，否则一般应有交集）
		if !linesAreCoverRegions(a, countA, b, countB) {
			im.Set(base.E, base.B, '1')
		}
	}

	// 线的外部和面的内部，必然有交集
	im.Set(base.E, base.I, '2')
}

// 线的外部，是否与面的边界有交集
func linesAreCoverRegions(a Geo1, ca int, b Geo2, cb int) bool {
	for i := 0; i < ca; i++ {
		region := b.GetSubRegion(i)
		for _, ring := range region {
			cover := false
			for j := 0; j < cb; j++ {
				line := a.GetSubLine(j)
				if line.IsCovers(algorithm.Line(ring)) {
					cover = true
					break
				}
			}
			if !cover {
				return false
			}
		}
	}
	return true
}

// 线的内部，与面的边界是否有交集，维度是0 or 1；没有交集返回-1
func linesIBRegions(a Geo1, ca int, b Geo2, cb int) (dim int) {
	dim = -1
	for i := 0; i < ca; i++ {
		line := a.GetSubLine(i)
		for j := 0; j < cb; j++ {
			region := b.GetSubRegion(j)
			for _, ring := range region {
				d := algorithm.Ring(ring).CalcBIDim(line)
				dim = base.IntMax(d, dim)
				if dim == 1 {
					return
				}
			}
		}
	}
	return
}

// 多面是否cover多线
func regionsIsCoversLines(a Geo2, ca int, b Geo1, cb int) bool {
	for i := 0; i < cb; i++ {
		line := b.GetSubLine(i)
		cover := false
		for j := 0; j < ca; j++ {
			region := a.GetSubRegion(j)
			if region.IsCoversLine(line) {
				cover = true
				break
			}
		}
		if !cover {
			return false
		}
	}
	return true
}

// 点集和面集的关系，有一个符合就ok
func pntsWithRegions(pnts []base.Point2D, geo Geo2, count int) (in, on, out bool) {
	for _, pnt := range pnts {
		var in0, on0, out0 bool
		for i := 0; i < count; i++ {
			region := geo.GetSubRegion(i)
			if !in0 && region.IsContainsPnt(pnt) {
				in0 = true
			} else if !on0 && region.IsTouchesPnt(pnt) {
				on0 = true
			} else if !out {
				out0 = true
			}
			if in0 && on0 && out0 {
				break
			}
		}
		if in0 {
			in = true
		}
		if on0 {
			on = true
		}
		if out0 {
			out = true
		}
	}
	return
}

// 线和线
func (this *GeoRelation) calc11(im *base.D9IM) {
	// 先看端点的情况
	a := this.A.(Geo1)
	b := this.B.(Geo1)
	pntsA := a.GetPntsB()
	pntsB := b.GetPntsB()
	// 边界
	if this.needBB() && pntsAreOverlap(pntsA, pntsB) {
		im.Set(base.B, base.B, '0')
	}

	// 线A的端点 与 线B的内部和外部
	if this.needBI() || this.needBE() {
		in, out := pntsWithLines(pntsA, b, this.B.SubCount())
		if in {
			im.Set(base.B, base.I, '0')
		}
		if out {
			im.Set(base.B, base.E, '0')
		}
	}

	// 线B的端点 与 线A的内部和外部
	if this.needIB() || this.needEB() {
		in, out := pntsWithLines(pntsB, a, this.A.SubCount())
		if in {
			im.Set(base.I, base.B, '0')
		}
		if out {
			im.Set(base.E, base.B, '0')
		}
	}

	// 看一下相互覆盖的关系
	countA := this.A.SubCount()
	countB := this.B.SubCount()
	if this.needEI() {
		for i := 0; i < countB; i++ {
			if !linesCoverLine(a, countA, b.GetSubLine(i)) {
				im.Set(base.E, base.I, '1')
				break
			}
		}
	}
	if this.needIE() {
		for i := 0; i < countA; i++ {
			if !linesCoverLine(b, countB, a.GetSubLine(i)) {
				im.Set(base.I, base.E, '1')
				break
			}
		}
	}

	if this.needII() {
		// 最后判断内部是否相交，交集为点还是线
		dim := linesIILines(a, countA, b, countB)
		if dim == 0 {
			im.Set(base.I, base.I, '0')
		} else if dim == 1 {
			im.Set(base.I, base.I, '1')
		}
	}
}

// 线和线内部的关系，返回交集的维度，没有交集返回-1
func linesIILines(lines1 Geo1, count1 int, lines2 Geo1, count2 int) (dim int) {
	dim = -1
	for i := 0; i < count1; i++ {
		for j := 0; j < count2; j++ {
			a := lines1.GetSubLine(i)
			b := lines1.GetSubLine(j)
			dim = base.IntMax(dim, a.CalcIIDim(b))
			if dim == 1 {
				return
			}
		}
	}
	return
}

func linesCoverLine(lines Geo1, count int, line algorithm.Line) bool {
	for i := 0; i < count; i++ {
		sub := lines.GetSubLine(i)
		if sub.IsCovers(line) {
			return true
		}
	}
	return false
}

// 判断点集和多线的关系
func pntsWithLines(pnts []base.Point2D, lines Geo1, count int) (in, out bool) {
	for _, v := range pnts {
		for i := 0; i < count; i++ {
			line := lines.GetSubLine(i)
			if !in && line.IsContainsPnt(v) {
				in = true
			} else if !out && !line.IsCoversPnt(v) {
				out = true
			}
			if in && out {
				return
			}
		}
	}
	return
}

// 点和面
func (this *GeoRelation) calc02(im *base.D9IM) {
	// 点的外部肯定和面的内部、边界都有交集
	im.Set(base.E, base.I, '2')
	im.Set(base.E, base.B, '1')

	// 剩下的就是看点是否在面的内部、边界和外部了
	a := this.A.(Geo0)
	b := this.B.(Geo2)
	pntsA := a.GetPnts()
	count := this.B.SubCount()
	if this.needII() || this.needIB() || this.needIE() {
		in, on, out := pntsWithRegions(pntsA, b, count)
		if in {
			im.Set(base.I, base.I, '0')
		}
		if on {
			im.Set(base.I, base.B, '0')
		}
		if out {
			im.Set(base.I, base.E, '0')
		}
	}
}

// 点和线
func (this *GeoRelation) calc01(im *base.D9IM) {
	// 点的外部肯定和线的内部、边界都有交集
	im.Set(base.E, base.I, '1')
	im.Set(base.E, base.B, '0')

	//剩下的就是 ：点是否在线的端点、内部和外部了
	var on, in, out bool
	a := this.A.(Geo0)
	b := this.B.(Geo1)
	pntsA := a.GetPnts()
	count := this.B.SubCount()
	if this.needIB() || this.needII() || this.needIE() {
		for i := 0; i < count; i++ {
			line := b.GetSubLine(i)
			for _, pnt := range pntsA {
				if !on && (pnt == line[0] || pnt == line[len(line)-1]) {
					// 点是否在线段端点
					im.Set(base.I, base.B, '0')
					on = true
				} else if !in && line.IsContainsPnt(pnt) {
					// 点是否在线段内部
					im.Set(base.I, base.I, '0')
					in = true
				} else if !out {
					// 点是否在线段外部
					im.Set(base.I, base.E, '0')
					out = true
				}
				if in && on && out {
					return
				}
			}
		}
	}
}

// 两个点
func (this *GeoRelation) calc00(im *base.D9IM) {
	// 0维对象没有边界，故而带边界的，都是 *
	im.Set(base.I, base.B, '*')
	im.Set(base.E, base.B, '*')

	// 点只需要判断双方的内部点，是否重合，以及是否都重合即可
	a := this.A.(Geo0)
	b := this.B.(Geo0)
	pntsA := a.GetPnts()
	pntsB := b.GetPnts()
	if this.needII() && pntsAreOverlap(pntsA, pntsB) {
		im.Set(base.I, base.I, '0')
	}
	if this.needIE() {
		for _, v := range pntsA {
			if pntIsOut(v, pntsB) {
				im.Set(base.I, base.E, '0')
				break
			}
		}
	}
	if this.needEI() {
		for _, v := range pntsB {
			if pntIsOut(v, pntsA) {
				im.Set(base.E, base.I, '0')
				break
			}
		}
	}
}

// 点是否在点集中
func pntIsWithin(pnt base.Point2D, pnts []base.Point2D) bool {
	for _, v := range pnts {
		if v == pnt {
			return true
		}
	}
	return false
}

// 点是否有重合，有一个就OK
func pntsAreOverlap(pnts1 []base.Point2D, pnts2 []base.Point2D) bool {
	for _, v := range pnts1 {
		for _, vv := range pnts2 {
			if v == vv {
				return true
			}
		}
	}
	return false
}

// 点是否不在点集中
func pntIsOut(pnt base.Point2D, pnts []base.Point2D) bool {
	for _, v := range pnts {
		if pnt != v {
			return true
		}
	}
	return false
}

func (this *GeoRelation) setDisjoint(im *base.D9IM) {
	im.Set(base.I, base.E, dimb2im(this.A.Dim()))
	im.Set(base.B, base.E, dimb2im(this.A.DimB()))
	im.Set(base.E, base.I, dimb2im(this.B.Dim()))
	im.Set(base.E, base.B, dimb2im(this.B.DimB()))
}

// 维度信息转化为im中的 0/1/2/F/*
// 结果维度的转化，-1为F
func dim2im(dim int) byte {
	if dim >= 0 {
		return strconv.Itoa(dim)[0]
	}
	// -1返回 F
	return 'F'
}

// 边界维度的转化，-1为*
func dimb2im(dim int) byte {
	if dim >= 0 {
		return strconv.Itoa(dim)[0]
	}
	// -1返回 *
	return '*'
}
