package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/algorithm"
	"gogis/base"
	"gogis/draw"
	"io"
)

func init() {
	RegisterGeo(TGeoPolygon, NewGeoPolygon)
}

func NewGeoPolygon() Geometry {
	return new(GeoPolygon)
}

type GeoPolygon struct {
	Points [][][]base.Point2D // []: 点串 [][]: 带洞的单个多边形 [][][] 带岛的多边形
	BBox   base.Rect2D
	GeoID
}

func (this *GeoPolygon) Type() GeoType {
	// if len(this.Points) == 1 {
	// 	return TGeoPolygon
	// } else if len(this.Points) > 1 {
	// 	return TGeoMultiPolygon
	// }
	return TGeoPolygon
}

func (this *GeoPolygon) Clone() Geometry {
	res := new(GeoPolygon)
	res.id = this.id
	res.BBox = this.BBox
	res.Points = make([][][]base.Point2D, len(this.Points))
	for i, v := range this.Points {
		res.Points[i] = make([][]base.Point2D, len(v))
		for ii, vv := range v {
			res.Points[i][ii] = make([]base.Point2D, len(vv))
			for iii, vvv := range vv {
				res.Points[i][ii][iii] = vvv
			}
		}
	}
	return res
}

func (this *GeoPolygon) GetBounds() base.Rect2D {
	return this.BBox
}

func (this *GeoPolygon) Dim() int {
	return 2
}

func (this *GeoPolygon) DimB() int {
	return 1
}

func (this *GeoPolygon) SubCount() int {
	return len(this.Points)
}

func (this *GeoPolygon) GetSubRegion(num int) algorithm.Region {
	return this.Points[num]
}

func (this *GeoPolygon) ComputeBounds() base.Rect2D {
	this.BBox.Init()
	for _, polygon := range this.Points {
		// 只有第一个是岛，需要计算bounds；后面有也是洞，可以不理会
		if len(polygon) >= 1 {
			this.BBox = this.BBox.Union(base.ComputeBounds(polygon[0]))
		}
	}
	return this.BBox
}

// 面的绘制
func (this *GeoPolygon) Draw(canvas *draw.Canvas) {
	for _, v := range this.Points {
		var geo = new(draw.Polygon)
		geo.Points = make([][]draw.Point, 0)
		for _, vv := range v {
			pnts := canvas.Params.Forwards(vv)
			geo.Points = append(geo.Points, pnts)
		}
		canvas.DrawPolyPolygon(geo)
	}
}

func (this *GeoPolygon) ConvertPrj(prjc *base.PrjConvert) {
	if prjc != nil {
		for i, v := range this.Points {
			for ii, vv := range v {
				this.Points[i][ii] = prjc.DoPnts(vv)
			}
		}
	}
	this.ComputeBounds()
}

func (this *GeoPolygon) Thin(dis2, angle float64) Geometry {
	var newgeo GeoPolygon
	newgeo.BBox = this.BBox
	newgeo.id = this.id
	newgeo.Points = make([][][]base.Point2D, 0, len(this.Points))
	for i, _ := range this.Points {
		pntss := make([][]base.Point2D, 0, len(this.Points[i]))
		for _, v := range this.Points[i] {
			pnts := algorithm.Ring(v).Thin(dis2, angle)
			if len(pnts) >= 3 {
				pntss = append(pntss, pnts)
			}
		}
		if len(pntss) > 0 {
			newgeo.Points = append(newgeo.Points, pntss)
		}
	}
	if len(newgeo.Points) == 0 {
		// newgeo.Make(this.BBox)
		return nil
	}
	return &newgeo
}

func (this *GeoPolygon) Make(bbox base.Rect2D) {
	this.BBox = bbox
	this.Points = make([][][]base.Point2D, 1)
	this.Points[0] = make([][]base.Point2D, 1)
	this.Points[0][0] = make([]base.Point2D, 5)
	copy(this.Points[0][0][0:4], bbox.ToPoints(false))
	this.Points[0][0][4] = this.Points[0][0][0]
}

// ================================================================ //

// wkb:
// WKBPolygon{
// byte  byteOrder;                               //字节序
// static  uint32  wkbType= 3;                    //几何体类型
// uint32  numRings;                                   //线串的个数
// LinearRing rings[numRings]}           //线串（环）的数组

// WKBMultiPolygon{
// 	bytebyteOrder;                                        //字节序
// 	staticuint32 wkbType = 6;                         //几何体类型
// 	uint32numPolygons;                                //多边形数目
// 	WKBPolygonpolygons[numPolygons]}      //多边形数组

func (this *GeoPolygon) From(data []byte, mode GeoMode) (result bool) {
	this.Points = this.Points[0:0] // 清空
	buf := bytes.NewBuffer(data)

	switch mode {
	case WKB:
		result = this.fromWKB(buf)
		this.ComputeBounds()
	// case WKT: // todo
	case GAIA:
		result = this.fromGAIA(buf)
	case Shape:
		return this.fromShp(buf)
	}

	return
}

func (this *GeoPolygon) fromShp(r io.Reader) bool {
	bbox, numParts, numPoints := loadShpOnePolyHeader(r)
	this.BBox = bbox

	parts := make([]int32, numParts+1)
	for i := int32(0); i < numParts; i++ {
		binary.Read(r, binary.LittleEndian, &parts[i])
	}
	parts[numParts] = numPoints

	this.Points = make([][][]base.Point2D, numParts)
	for i := int32(0); i < numParts; i++ {
		this.Points[i] = make([][]base.Point2D, 1)
		this.Points[i][0] = make([]base.Point2D, parts[i+1]-parts[i])
		binary.Read(r, binary.LittleEndian, this.Points[i][0])
	}
	return true
}

func (this *GeoPolygon) fromGAIA(r io.Reader) bool {
	var begin byte
	binary.Read(r, binary.LittleEndian, &begin)
	if begin == byte(0X00) {
		var info GAIAInfo
		byteOrder := info.From(r)
		this.BBox = info.bbox

		var geoType int32
		binary.Read(r, byteOrder, &geoType)
		if geoType == 6 {
			// int32 numPolygon; //子对象个数
			// PolygonEntity[] polygons[numPolygon]; //子对象数据
			var subCount int32
			binary.Read(r, byteOrder, &subCount)
			this.Points = make([][][]base.Point2D, subCount)
			for i := int32(0); i < subCount; i++ {
				// PolygonEntity {
				// 	static byte gaiaEntityMark = 0x69; //子对象标识
				// 	PolygonData data; //子对象数据
				// 	}
				var mark byte
				binary.Read(r, byteOrder, &mark)
				if mark == 0x69 {
					this.Points[i] = gaia2Polygon(r, byteOrder)
				}
			}
			return true
		}
	}
	return false
}

func (this *GeoPolygon) fromWKB(r io.Reader) bool {
	var order byte
	binary.Read(r, binary.LittleEndian, &order)
	byteOrder := WkbByte2Order(order)
	var wkbType WkbGeoType
	binary.Read(r, byteOrder, &wkbType)
	if wkbType == WkbPolygon { // 简单面
		this.Points = append(this.Points, Bytes2WkbPolygon(r, byteOrder))
	} else if wkbType == WkbMultiPolygon { // 复杂面
		var numPolygons uint32
		binary.Read(r, byteOrder, &numPolygons)
		this.Points = make([][][]base.Point2D, numPolygons)
		for i := uint32(0); i < numPolygons; i++ {
			var order byte
			binary.Read(r, binary.LittleEndian, &order)
			byteOrder := WkbByte2Order(order)
			var wkbType WkbGeoType
			binary.Read(r, byteOrder, &wkbType)
			this.Points[i] = Bytes2WkbPolygon(r, byteOrder)
		}
	}
	return true
}

func (this *GeoPolygon) To(mode GeoMode) []byte {
	switch mode {
	case WKB:
		return this.toWkb()
	case WKT:
		// todo
	}
	return nil
}

func (this *GeoPolygon) toWkb() []byte {
	var buf bytes.Buffer
	if len(this.Points) == 1 { // 简单面
		WkbPolygon2Bytes(this.Points[0], &buf)
	} else { // 复杂面
		binary.Write(&buf, binary.LittleEndian, byte(WkbLittle))
		binary.Write(&buf, binary.LittleEndian, uint32(WkbMultiPolygon))
		binary.Write(&buf, binary.LittleEndian, uint32(len(this.Points)))
		for _, v := range this.Points {
			WkbPolygon2Bytes(v, &buf)
		}
	}
	return buf.Bytes()
}

// ================================================================ //

func (this *GeoPolygon) IsRelate(mode base.SpatialMode, rect base.Rect2D) bool {
	switch mode {
	case base.BBoxIntersects:
		return this.BBox.IsIntersects(rect)
	case base.Intersects, base.Undefined:
		return this.IsIntersects(rect)
	case base.Disjoint:
		return !this.IsIntersects(rect)
	case base.Within:
		// 在rect内部，且边界不接触
		return rect.IsContains(this.BBox)
	case base.CoveredBy:
		// 在rect的内部（且边界可接触）
		return rect.IsCovers(this.BBox)
	case base.Touches:
		return this.IsTouches(rect)
	case base.Crosses:
		// 面无法cross面，返回false
		return false
	case base.Equals:
		return this.IsEquals(rect)
		// todo
		// var polygon GeoPolygon
		// polygon.Make(rect)
		// return this.IsEquals(&polygon)
		// return false
	case base.Contains:
		return this.IsContains(rect)
	case base.Covers:
		return this.IsCovers(rect)
	case base.Overlaps:
		return this.IsOverlaps(rect)
	default:
		var im base.D9IM
		if im.Init(string(mode)) {
			imRes := this.CalcRelateIM(rect)
			return imRes.MatchIM(im)
		}
	}
	return false
}

// 计算点和rect的九交模型矩阵；
func (this *GeoPolygon) CalcRelateIM(rect base.Rect2D) (im base.D9IM) {
	// 初始化为 disjoint
	im.Init("FFFFFFFF*")
	// 如果有线段和rect相交，那么II/IB/BI/BB 都能确定
	for _, v := range this.Points {
		for _, vv := range v {
			if algorithm.Line(vv).IsCrossesRect(rect) {
				im.Set(base.I, base.I, '2')
				im.Set(base.I, base.B, '1')
				im.Set(base.B, base.I, '1')
				im.Set(base.B, base.B, '0')
				goto EndBI
			}
		}
	}
EndBI:

	// 如果点在rect内,则说明 II=2,BI=1
	if im.Get(base.I, base.I) != '2' && im.Get(base.B, base.I) != '1' {
		for _, v := range this.Points {
			for _, vv := range v {
				for i := 0; i < len(vv); i++ {
					if rect.IsContainsPnt(vv[i]) {
						im.Set(base.I, base.I, '2')
						im.Set(base.B, base.I, '1')
						goto EndII
					}
				}
			}
		}
	}
EndII:

	// 如果rect的某个角点在面内部，II=2，IB=1
	if im.Get(base.I, base.I) != '2' && im.Get(base.I, base.B) != '1' {
		cPnts := rect.ToPoints(false)
		for _, pnt := range cPnts {
			for _, v := range this.Points {
				if algorithm.Region(v).IsContainsPnt(pnt) {
					im.Set(base.I, base.I, '2')
					im.Set(base.I, base.B, '1')
					goto EndIB
				}
			}
		}
	}
EndIB:

	// 再看两个的边界是否重叠（1维）
	for _, v := range this.Points {
		for _, vv := range v {
			for i := 1; i < len(vv); i++ {
				if (algorithm.Segment{P1: vv[i-1], P2: vv[i]}).IsTouchesRect(rect) {
					im.Set(base.B, base.B, '1')
					goto EndBB
				}
			}
		}
	}
EndBB:

	// 如果不能覆盖rect，则说明自己的外部和rect内部、边界均有交集
	if !this.IsCovers(rect) {
		im.Set(base.E, base.I, '2')
		im.Set(base.E, base.B, '1')
	}
	// 如果自己不被rect覆盖，说明自己的内部和边界，与rect的外部有交集
	if !rect.IsCovers(this.BBox) {
		im.Set(base.I, base.E, '2')
		im.Set(base.B, base.E, '1')
	}
	return
}

func (this *GeoPolygon) IsEquals(rect base.Rect2D) bool {
	if !this.BBox.IsEquals(rect) {
		return false
	}
	// 只有一个子对象，且不能有洞
	if len(this.Points) != 1 || len(this.Points[0]) != 1 {
		return false
	}
	// 各自的节点都在对方的边界上
	shell := this.Points[0][0]
	for _, pnt := range shell {
		if !rect.IsTouchesPnt(pnt) {
			return false
		}
	}
	cPnts := rect.ToPoints(false)
	for _, pnt := range cPnts {
		if !algorithm.Line(shell).IsCoversPnt(pnt) {
			return false
		}
	}
	return true
}

// 内部和内部部分相交
func (this *GeoPolygon) IsOverlaps(rect base.Rect2D) bool {
	// 和一个子对象overlap，那就ok
	for _, v := range this.Points {
		if algorithm.Region(v).IsOverlapsRect(rect) {
			return true
		}
	}
	// 若有多个子对象，有在rect内的，还有在rect外的，也是overlap
	var in, out bool
	for _, v := range this.Points {
		bbox := base.ComputeBounds(v[0])
		if bbox.IsCovers(rect) {
			in = true
		} else {
			out = true
		}
		if in && out {
			return true
		}
	}
	return false
}

// 内部不能有交集,但是内部与边界，边界与边界有交集
func (this *GeoPolygon) IsTouches(rect base.Rect2D) bool {
	// bbox 必然有交集
	if !this.subBboxesIsIntersects(rect) {
		return false
	}

	touchOne := false
	for _, v := range this.Points {
		// 有一个子对象和rect touch
		if !touchOne && algorithm.Region(v).IsTouchesRect(rect) {
			touchOne = true
		}
		// 且所有的子对象都不能和rect内部有交集（overlaps）
		if algorithm.Region(v).IsOverlapsRect(rect) {
			return false
		}
	}
	return touchOne
}

// 自己的bbox（含子对象的），是否存在被bbox包裹的情况；
// 有一个就ok；用在拉框查询时
func (this *GeoPolygon) bboxesIsCoveredBy(bbox base.Rect2D) bool {
	if bbox.IsCovers(this.BBox) {
		return true
	}
	// 如果有多个子对象，则每个都要单独判断；有一个就ok
	if len(this.Points) > 1 {
		for _, v := range this.Points {
			// 这里只需要判断外环；内环是洞，无需理会
			subBbox := base.ComputeBounds(v[0])
			if bbox.IsCovers(subBbox) {
				return true
			}
		}
	}
	return false
}

// 子对象的bbox，是否与bbox相交；有一个相交就返回true
// 没有子对象，就用自己的
func (this *GeoPolygon) subBboxesIsIntersects(bbox base.Rect2D) bool {
	if len(this.Points) > 1 {
		for _, v := range this.Points {
			subBbox := base.ComputeBounds(v[0])
			if subBbox.IsIntersects(bbox) {
				return true
			}
		}
	} else {
		if this.BBox.IsIntersects(bbox) {
			return true
		}
	}
	return false
}

// 判断是否与rect相交
func (this *GeoPolygon) IsIntersects(rect base.Rect2D) bool {
	// 先用bbox是否cover进行判断
	if this.bboxesIsCoveredBy(rect) {
		return true
	}
	// 子对象的bbox都没有一个相交的，肯定不会相交了
	if !this.subBboxesIsIntersects(rect) {
		return false
	}

	// 如果rect有任意一个顶点在一个子对象的外环之内（边界亦可），
	// 且bbox整体不在该子对象的某个内环之内（边界不可），返回true
	pnts := rect.ToPoints(false)
	for _, v := range this.Points {
		oneInPolygon := false
		for _, pnt := range pnts {
			if algorithm.Ring(v[0]).IsCoversPnt(pnt) {
				oneInPolygon = true
				break
			}
		}
		if oneInPolygon {
			for j := 1; j < len(v); j++ {
				if algorithm.Ring(v[j]).IsContainsPnts(pnts) {
					oneInPolygon = false
					break
				}
			}
		}
		if oneInPolygon {
			return true
		}
	}

	return false
}

func (this *GeoPolygon) IsCovers(rect base.Rect2D) bool {
	// 先用bbox是否cover进行判断
	if !this.bboxesIsCoveredBy(rect) {
		return false
	}
	// 子对象的bbox都没有一个相交的，肯定没关系了
	if !this.subBboxesIsIntersects(rect) {
		return false
	}
	// 有一个子对象cover，就ok
	for _, v := range this.Points {
		if algorithm.Region(v).IsCoversRect(rect) {
			return true
		}
	}
	return false
}

func (this *GeoPolygon) IsContains(rect base.Rect2D) bool {
	// 首先必须cover
	if this.IsCovers(rect) {
		// 然后角点不能在边界上，外环内环都不行
		cPnts := rect.ToPoints(false)
		for _, v := range this.Points {
			for _, vv := range v {
				for _, pnt := range cPnts {
					if algorithm.Line(vv).IsCoversPnt(pnt) {
						return false
					}
				}
			}
		}
		return true
	}
	return false
}

// ================================================================ //
/*
func (this *GeoPolygon) IsRelated(mode base.SpatialMode, geo Geometry) bool {
	return false
}

// todo
func (this *GeoPolygon) CalcRelateIM(geo Geometry) (im base.D9IM) {
	im.Init("FFFFFFFFF")
	return
}

func (this *GeoPolygon) IsEquals(geo Geometry) bool {
	// 类型必须一样
	if geoPolygon, ok := geo.(*GeoPolygon); ok {
		// 边框必须一致
		if this.BBox != geoPolygon.BBox {
			return false
		}
		// 子对象个数必须 相等
		if len(this.Points) != len(geoPolygon.Points) {
			return false
		}
		// 对比每个子对象；允许子对象顺序调换
		used := make([]bool, len(geoPolygon.Points))
		for i, v := range this.Points {
			subBbox1 := base.ComputeBounds(v[0])
			for ii, vv := range geoPolygon.Points {
				if !used[ii] { // 没用过的才进行对比
					subBbox2 := base.ComputeBounds(vv[0])
					// 用bbox判断是否为对应的子对象，不能保证完全准确
					if subBbox1 == subBbox2 {
						if !algorithm.Region(v).IsEquals(algorithm.Region(vv)) {
							return false
						}
						used[i] = true
					}
				}
			}
		}
	}
	return true
}
*/
