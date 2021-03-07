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
	RegisterGeo(TGeoPolyline, NewGeoPolyline)
}

func NewGeoPolyline() Geometry {
	return new(GeoPolyline)
}

type GeoPolyline struct {
	Points [][]base.Point2D
	BBox   base.Rect2D
	GeoID
	// vao uint32
}

func (this *GeoPolyline) Type() GeoType {
	// 	if len(this.Points) == 1 {
	// 		return TGeoLineString
	// 	} else if len(this.Points) > 1 {
	// 		return TGeoMultiLineString
	// 	}
	return TGeoPolyline
}

func (this *GeoPolyline) Clone() Geometry {
	res := new(GeoPolyline)
	// res.vao = this.vao
	res.id = this.id
	res.BBox = this.BBox
	res.Points = make([][]base.Point2D, len(this.Points))
	for i, v := range this.Points {
		res.Points[i] = make([]base.Point2D, len(v))
		for ii, vv := range v {
			res.Points[i][ii] = vv
		}
	}
	return res
}

func (this *GeoPolyline) GetBounds() base.Rect2D {
	return this.BBox
}

func (this *GeoPolyline) Dim() int {
	return 1
}

func (this *GeoPolyline) DimB() int {
	return 1
}

func (this *GeoPolyline) SubCount() int {
	return len(this.Points)
}

func (this *GeoPolyline) GetPntsB() (pnts []base.Point2D) {
	pnts = make([]base.Point2D, len(this.Points)*2)
	for i, v := range this.Points {
		pnts[i*2] = v[0]
		pnts[i*2+1] = v[len(v)-1]
	}
	return
}

func (this *GeoPolyline) GetSubLine(num int) algorithm.Line {
	return this.Points[num]
}

// func (this *GeoPolyline) GetPnts(ibe base.IBE) (pnts [][]base.Point2D) {
// 	switch ibe {
// 	case  base.I:
// 	case  base.B:
// 		pnts = make([][]base.Point2D, 1)
// 		pnts[0] = make([]base.Point2D, 0, len(this.Points)*2)
// 		for _,v:=range this.Points {
// 			pnts[0] = append(pnts[0], v[0])
// 			pnts[0] = append(pnts[0], v[len(v)-1])
// 		}
// 	}
// 	return
// }

func (this *GeoPolyline) ComputeBounds() base.Rect2D {
	this.BBox.Init()
	for _, points := range this.Points {
		this.BBox = this.BBox.Union(base.ComputeBounds(points))
	}
	return this.BBox
}

func (this *GeoPolyline) Draw(canvas draw.Canvas) {
	// var line = new(draw.Polyline)
	// line.Points = make([][]draw.Point, len(this.Points))
	for _, v := range this.Points {
		canvas.DrawLine(canvas.Forwards(v))
	}
	// canvas.DrawPolyPolyline(line)
}

// func (this *GeoPolyline) Draw(canvas *draw.Canvas) {
// 	if this.vao == 0 {
// 		points := make([]float32, 0)
// 		for _, v := range this.Points {
// 			points = append(points, canvas.Params.Forward32s(v)...)
// 		}
// 		// canvas.DrawPolyPolyline(line)
// 		this.vao = draw.MakeVao(points)
// 	}
// 	count := 0
// 	for _, v := range this.Points {
// 		count += len(v)
// 	}
// 	canvas.DrawLineVao(this.vao, int32(count))
// }

func (this *GeoPolyline) ConvertPrj(prjc *base.PrjConvert) {
	if prjc != nil {
		for i, v := range this.Points {
			this.Points[i] = prjc.DoPnts(v)
		}
	}
	this.ComputeBounds()
}

// ================================================================ //

func (this *GeoPolyline) Thin(dis2, angle float64) Geometry {
	var newgeo GeoPolyline
	newgeo.id = this.id
	newgeo.BBox = this.BBox
	newgeo.Points = make([][]base.Point2D, 0, len(this.Points))
	for _, v := range this.Points {
		pnts := algorithm.Line(v).Thin(dis2, angle)
		if pnts != nil {
			newgeo.Points = append(newgeo.Points, pnts)
		}
	}
	if len(newgeo.Points) == 0 {
		return nil
	}
	return &newgeo
}

// ================================================================ //

// wkb:
// WKBLineString
// byte  byteOrder;                               //字节序
// static  uint32  wkbType = 2;                    //几何体类型
// uint32  numPoints;                                  //点的个数
// Point  points[numPoints]}                 //点的坐标数组

// WKBMultiLineString
// byte  byteOrder;                                                    //字节序
// staticuint32 wkbType = 5;                                       //几何体类型
// uint32  numLineStrings;                                        //线串的个数
// WKBLineString  lineStrings[numLineStrings]}         //线串数组

func (this *GeoPolyline) From(data []byte, mode GeoMode) bool {
	this.Points = this.Points[0:0] // 清空
	buf := bytes.NewBuffer(data)

	switch mode {
	case WKB:
		return this.fromWkb(buf)
	case GAIA:
		return this.fromGAIA(buf)
	// case WKT: // todo
	case Shape:
		return this.fromShp(buf)
	}
	return false
}

func (this *GeoPolyline) fromWkb(r io.Reader) bool {
	var order byte
	binary.Read(r, binary.LittleEndian, &order)
	byteOrder := WkbByte2Order(order)
	var wkbType WkbGeoType
	binary.Read(r, byteOrder, &wkbType)
	if wkbType == WkbLineString { // 简单线
		this.Points = append(this.Points, Bytes2WkbLinearRing(r, byteOrder))
	} else if wkbType == WkbMultiLineString { // 复杂线
		var numLineStrings uint32
		binary.Read(r, byteOrder, &numLineStrings)
		this.Points = make([][]base.Point2D, numLineStrings)
		for i := uint32(0); i < numLineStrings; i++ {
			var order byte
			binary.Read(r, binary.LittleEndian, &order)
			byteOrder := WkbByte2Order(order)
			var wkbType WkbGeoType
			binary.Read(r, byteOrder, &wkbType)
			this.Points[i] = Bytes2WkbLinearRing(r, byteOrder)
		}
	}
	this.ComputeBounds()
	return true
}

func (this *GeoPolyline) fromGAIA(r io.Reader) bool {
	var begin byte
	binary.Read(r, binary.LittleEndian, &begin)
	if begin == byte(0X00) {
		var info GAIAInfo
		byteOrder := info.From(r)
		this.BBox = info.bbox

		var geoType int32
		binary.Read(r, byteOrder, &geoType)
		if geoType == 5 {
			var subCount int32
			binary.Read(r, byteOrder, &subCount)
			this.Points = make([][]base.Point2D, subCount)
			for i := int32(0); i < subCount; i++ {
				// PolygonEntity {
				// 	static byte gaiaEntityMark = 0x69; //子对象标识
				// 	PolygonData data; //子对象数据
				// 	}
				this.Points[i] = gaia2LineString(r, byteOrder)
			}
			return true
		}
	}
	return false
}

// type shpPolyline struct {
// 	shpType                  // 图形类型，==3
// 	bbox       base.Rect2D    // 当前线状目标的坐标范围
// 	numParts  int32          // 当前线目标所包含的子线段的个数
// 	numPoints int32          // 当前线目标所包含的顶点个数
// 	parts     []int32        // 每个子线段的第一个坐标点在 Points 的位置
// 	points    []base.Point2D // 记录所有坐标点的数组
// }
func (this *GeoPolyline) fromShp(r io.Reader) bool {
	// var polyline geometry.GeoPolyline
	bbox, numParts, numPoints := loadShpOnePolyHeader(r)
	this.BBox = bbox

	parts := make([]int32, numParts, numParts+1)
	binary.Read(r, binary.LittleEndian, parts)
	parts = append(parts, numPoints) // 最后增加一个，方便后面的计算

	this.Points = make([][]base.Point2D, numParts)
	for i := int32(0); i < numParts; i++ {
		this.Points[i] = make([]base.Point2D, parts[i+1]-parts[i])
		binary.Read(r, binary.LittleEndian, this.Points[i])
	}
	return true
}

func (this *GeoPolyline) To(mode GeoMode) []byte {
	switch mode {
	case WKB:
		return this.toWkb()
	case WKT:
		// todo
	case GAIA:
		return this.toGAIA()
	}
	return nil
}

func (this *GeoPolyline) toWkb() []byte {
	var buf bytes.Buffer
	if len(this.Points) == 1 { // 简单线
		WkbLineString2Bytes(this.Points[0], &buf)
	} else { // 复杂线
		binary.Write(&buf, binary.LittleEndian, byte(WkbLittle))
		binary.Write(&buf, binary.LittleEndian, uint32(WkbMultiLineString))
		binary.Write(&buf, binary.LittleEndian, uint32(len(this.Points)))
		for _, v := range this.Points {
			WkbLineString2Bytes(v, &buf)
		}
	}
	return buf.Bytes()
}

func (this *GeoPolyline) toGAIA() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, byte(0))
	var info GAIAInfo
	info.Init(this.BBox, 0)
	binary.Write(&buf, binary.LittleEndian, info.To())
	binary.Write(&buf, binary.LittleEndian, int32(5)) // type
	binary.Write(&buf, binary.LittleEndian, int32(len(this.Points)))
	for _, v := range this.Points {
		binary.Write(&buf, binary.LittleEndian, byte(0x69))
		binary.Write(&buf, binary.LittleEndian, int32(2))
		binary.Write(&buf, binary.LittleEndian, int32(len(v)))
		binary.Write(&buf, binary.LittleEndian, v)
	}
	binary.Write(&buf, binary.LittleEndian, byte(0xFE))
	return buf.Bytes()
}

// ======================================================== //

func (this *GeoPolyline) IsRelate(mode base.SpatialMode, rect base.Rect2D) bool {
	switch mode {
	case base.Intersects, base.Undefined:
		return this.IsIntersects(rect)
	case base.BBoxIntersects:
		return this.BBox.IsIntersects(rect)
	case base.Disjoint:
		return !this.IsIntersects(rect)
	case base.Touches:
		return this.IsTouches(rect)
	case base.Within:
		// 在rect内部，且边界不接触
		return rect.IsContains(this.BBox)
	case base.CoveredBy:
		// 在rect的内部（且边界可接触）
		return rect.IsCovers(this.BBox)
	case base.Equals, base.Contains, base.Covers, base.Overlaps:
		return false
	case base.Crosses:
		return this.IsCrosses(rect)
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
func (this *GeoPolyline) CalcRelateIM(rect base.Rect2D) (im base.D9IM) {
	// 初始化为 disjoint
	im.Init("FFFFFF***")

	// 先看 线的端点:
	var in, on, out bool
	for _, v := range this.Points {
		for _, pnt := range v {
			// 若有端点到rect内部，则 BI=0, 且II=1
			if !in && rect.IsContainsPnt(pnt) {
				in = true
				im.Set(base.B, base.I, '0')
				im.Set(base.I, base.I, '1')
			}
			// 若有端点在rect边界上，则BB=0
			if !on && rect.IsTouchesPnt(pnt) {
				on = true
				im.Set(base.B, base.B, '0')
			}
			// 若有端点在rect之外，则 BE=0, 且IE=1
			if !out && rect.IsDisjointPnt(pnt) {
				out = true
				im.Set(base.B, base.E, '0')
				im.Set(base.I, base.E, '1')
			}
			if in && on && out {
				break
			}
		}
		if in && on && out {
			break
		}
	}

	// 剩下线面IB需要仔细判断，交集有可能是点，也有可能是线
	var res0 bool
	for _, v := range this.Points {
		for i := 1; i < len(v); i++ {
			seg := algorithm.Segment{P1: v[i], P2: v[i-1]}
			if !res0 && seg.IsCrossesRect(rect) {
				res0 = true
				im.Set(base.I, base.B, '0')
			}
			// 若有1，则可以返回了
			if seg.IsTouchesRect(rect) {
				im.Set(base.I, base.B, '1')
				return
			}
		}
	}
	return
}

// 内部是否有交集
func (this *GeoPolyline) IsIIRect(rect base.Rect2D) bool {
	if !this.subBboxesIsIntersects(rect) {
		return false
	}
	// 只要有一个点在rect内，就ok
	for _, v := range this.Points {
		for _, pnt := range v {
			if rect.IsContainsPnt(pnt) {
				return true
			}
		}
	}
	return false
}

func (this *GeoPolyline) IsCrosses(rect base.Rect2D) bool {
	for _, v := range this.Points {
		if algorithm.Line(v).IsCrossesRect(rect) {
			return true
		}
	}
	return false
}

// 挨着，但折线不能到rect内部
func (this *GeoPolyline) IsTouches(rect base.Rect2D) bool {
	// 边框还是要有交集的
	if !this.subBboxesIsIntersects(rect) {
		return false
	}
	// 必须有一个子对象和rect touch，但所有子对象都不能到rect内部（即线不能cross rect）
	for _, v := range this.Points {
		if algorithm.Line(v).IsCrossesRect(rect) {
			return false
		}
	}

	for _, v := range this.Points {
		if algorithm.Line(v).IsTouchesRect(rect) {
			return true
		}
	}
	return false
}

func (this *GeoPolyline) IsIntersects(rect base.Rect2D) bool {
	// 先用bbox是否被cover进行判断
	if this.bboxesIsCoveredBy(rect) {
		return true
	}
	// 子对象的bbox都没有一个相交的，肯定不会相交了
	if !this.subBboxesIsIntersects(rect) {
		return false
	}

	// 如果有任意一个点被 bbox所覆盖，就ok
	for _, v := range this.Points {
		for _, vv := range v {
			if rect.IsCoversPnt(vv) {
				return true
			}
		}
	}

	// 仍然没有判断出来，就必须看看每个线段是否与bbox相交了
	for _, v := range this.Points {
		if algorithm.Line(v).IsIntersectsRect(rect) {
			return true
		}
	}

	return false
}

// 自己的bbox（含子对象的），是否存在被bbox包裹的情况；
// 有一个就ok；用在拉框查询时
func (this *GeoPolyline) bboxesIsCoveredBy(rect base.Rect2D) bool {
	if rect.IsCovers(this.BBox) {
		return true
	}
	// 如果有多个子对象，则每个都要单独判断；有一个就ok
	if len(this.Points) > 1 {
		for _, v := range this.Points {
			subBbox := base.ComputeBounds(v)
			if rect.IsCovers(subBbox) {
				return true
			}
		}
	}
	return false
}

// 子对象的bbox，是否与bbox相交；有一个相交就返回true
// 没有子对象，就用自己的
func (this *GeoPolyline) subBboxesIsIntersects(rect base.Rect2D) bool {
	if len(this.Points) > 1 {
		for _, v := range this.Points {
			subBbox := base.ComputeBounds(v)
			if subBbox.IsIntersects(rect) {
				return true
			}
		}
	} else {
		if this.BBox.IsIntersects(rect) {
			return true
		}
	}
	return false
}

// 前提是 bbox退化为一个点
// func (this *GeoPolyline) IsCoversRect(bbox base.Rect2D) bool {
// 	// 允许bbox在端点上
// 	if bbox.Max == bbox.Min && this.BBox.IsCoversPnt(bbox.Max) {
// 		for _, v := range this.Points {
// 			if !base.ComputeBounds(v).IsCoversPnt(bbox.Max) {
// 				return false
// 			}
// 			if algorithm.Line(v).IsCoversPnt(bbox.Max) {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// 前提是 bbox退化为一个点
// func (this *GeoPolyline) IsContainsRect(bbox base.Rect2D) bool {
// 	if this.IsCoversRect(bbox) {
// 		for _, v := range this.Points {
// 			// 不允许bbox在端点上
// 			if v[0] == bbox.Max || v[len(v)-1] == bbox.Max {
// 				return false
// 			}
// 		}
// 	}
// 	return false
// }

// ================================================================ //
/*
func (this *GeoPolyline) IsRelated(mode base.SpatialMode, geo Geometry) bool {
	switch mode {
	case base.Equals:
		return this.IsEquals(geo)
	case base.BBoxIntersects:
		return this.GetBounds().IsIntersects(geo.GetBounds())
	case base.Intersects, base.Undefined:
		return this.IsIntersects(geo)

		// todo
	case base.Disjoint:
		// return this.IsDisjoint(geo)
	case base.Touches:
		// return this.IsTouches(geo)
	case base.Within:
		// return this.IsWithin(geo)
	case base.CoveredBy:
		// return this.IsCoveredBy(geo)
	case base.Overlaps:
		// return this.IsOverlaps(geo)
	case base.Contains:
		// return this.IsContains(geo)
	case base.Covers:
		// return this.IsCovers(geo)
	case base.Crosses: // 点无法cross任何对象
		return false
	default:
		var im base.D9IM
		if im.Init(string(mode)) {
			imRes := this.CalcRelateIM(geo)
			return imRes.MatchIM(im)
		}
	}
	return false
}

// todo
func (this *GeoPolyline) CalcRelateIM(geo Geometry) (im base.D9IM) {
	im.Init("FFFFFFFFF")
	return
}

func (this *GeoPolyline) IsIntersects(geo Geometry) bool {
	switch geo.Type() {
	case TGeoPoint:
		if pnt, ok := geo.(*GeoPoint); ok {
			return this.IsIntersectsPnt(pnt)
		}
	case TGeoPolyline:
		if polyline, ok := geo.(*GeoPolyline); ok {
			return this.IsIntersectsPolyline(polyline)
		}
	case TGeoPolygon:
		if polygon, ok := geo.(*GeoPolygon); ok {
			return this.IsIntersectsPolygon(polygon)
		}
	}
	return false
}

// todo
func (this *GeoPolyline) IsIntersectsPolygon(polygon *GeoPolygon) bool {
	if this.BBox.IsDisjoint(this.BBox) {
		return false
	}
	// for _, v := range this.Points {
	// 	for _, vv := range polyline.Points {
	// 		if algorithm.Line(v).IsIntersects(vv) {
	// 			return true
	// 		}
	// 	}
	// }
	return false
}

func (this *GeoPolyline) IsIntersectsPolyline(polyline *GeoPolyline) bool {
	if this.BBox.IsDisjoint(this.BBox) {
		return false
	}
	for _, v := range this.Points {
		for _, vv := range polyline.Points {
			if algorithm.Line(v).IsIntersects(vv) {
				return true
			}
		}
	}
	return false
}

func (this *GeoPolyline) IsIntersectsPnt(geoPnt *GeoPoint) bool {
	for _, v := range this.Points {
		if algorithm.Line(v).IsCoversPnt(geoPnt.Point2D) {
			return true
		}
	}
	return false
}

func (this *GeoPolyline) IsEquals(geo Geometry) bool {
	// 类型必须一样 todo
	if geoPolyline, ok := geo.(*GeoPolyline); ok {
		// 边框必须一致
		if this.BBox != geoPolyline.BBox {
			return false
		}
		// 子对象个数必须 相等
		if len(this.Points) != len(geoPolyline.Points) {
			return false
		}
		// 对比每个子对象；允许子对象顺序调换
		used := make([]bool, len(geoPolyline.Points))
		for i, v := range this.Points {
			subBbox1 := base.ComputeBounds(v)
			for ii, vv := range geoPolyline.Points {
				if !used[ii] { // 没用过的才进行对比
					subBbox2 := base.ComputeBounds(vv)
					// 用bbox判断是否为对应的子对象，不能保证完全准确
					if subBbox1 == subBbox2 {
						if !algorithm.Line(v).IsEquals(vv) {
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
