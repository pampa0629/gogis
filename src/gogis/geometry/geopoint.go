package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
	"gogis/draw"
	"io"
)

func init() {
	RegisterGeo(TGeoPoint, NewGeoPoint)
}

func NewGeoPoint() Geometry {
	return new(GeoPoint)
}

type GeoPoint struct {
	base.Point2D
	GeoID
}

func (this *GeoPoint) Type() GeoType {
	return TGeoPoint
}

func (this *GeoPoint) GetBounds() base.Rect2D {
	return base.NewRect2D(this.X, this.Y, this.X, this.Y)
}

func (this *GeoPoint) Dim() int {
	return 0
}

func (this *GeoPoint) DimB() int {
	return -1
}

// 得到组成内部或边界的点串
// func (this *GeoPoint) GetPnts(ibe base.IBE) (pnts [][]base.Point2D) {
// 	if ibe == base.I {
// 		pnts = make([][]base.Point2D, 1)
// 		pnts[0] = make([]base.Point2D, 1)
// 		pnts[0][0] = this.Point2D
// 	}
// 	return
// }

func (this *GeoPoint) SubCount() int {
	return 1
}

func (this *GeoPoint) GetPnts() (pnts []base.Point2D) {
	pnts = make([]base.Point2D, 1)
	pnts[0] = this.Point2D
	return
}

func (this *GeoPoint) Clone() Geometry {
	res := new(GeoPoint)
	res.id = this.id
	res.Point2D = this.Point2D
	return res
}

func (this *GeoPoint) Draw(canvas draw.Canvas) {
	pnt := canvas.Forward(this.Point2D)
	canvas.DrawPoint(pnt)
}

func (this *GeoPoint) ConvertPrj(prjc *base.PrjConvert) {
	if prjc != nil {
		this.Point2D = prjc.DoPnt(this.Point2D)
	}
}

// ================================================================ //

// 点就只能返回自身
func (this *GeoPoint) Thin(dis2, angle float64) Geometry {
	return this
}

// ================================================================ //

func (this *GeoPoint) From(data []byte, mode GeoMode) bool {
	r := bytes.NewBuffer(data)
	switch mode {
	case WKB:
		return this.fromWKB(r)
	case WKT:
		// todo
	case GAIA:
		return this.fromGAIA(r)
	case Shape:
		return this.fromShp(r)
	}
	return false
}

func (this *GeoPoint) fromShp(r io.Reader) bool {
	var geotype int32
	binary.Read(r, binary.LittleEndian, &geotype)
	if geotype == 1 {
		binary.Read(r, binary.LittleEndian, &this.Point2D)
		return true
	}
	return false
}

func (this *GeoPoint) fromGAIA(r io.Reader) bool {
	var begin byte
	binary.Read(r, binary.LittleEndian, &begin)
	if begin == byte(0X00) {
		var info GAIAInfo
		byteOrder := info.From(r)
		var geoType int32
		binary.Read(r, byteOrder, &geoType)
		if geoType == 1 {
			binary.Read(r, byteOrder, &this.Point2D)
			var end byte
			binary.Read(r, binary.LittleEndian, &end)
			return true
		} else if geoType == 4 {
			// todo 暂时先这么处理，后续增加多点对象
			var count int32
			binary.Read(r, binary.LittleEndian, &count)
			for i := int32(0); i < count; i++ {
				var mark byte
				binary.Read(r, binary.LittleEndian, &mark)
				var geoType int32
				binary.Read(r, byteOrder, &geoType)
				binary.Read(r, byteOrder, &this.Point2D)
			}
		}
		return true
	}
	return false
}

// wkb:
// byte  byteOrder;                        //字节序 1
// static  uint32  wkbType= 1;             //几何体类型 4
// Point  point                             //点的坐标 8*2
func (this *GeoPoint) fromWKB(r io.Reader) bool {
	var order byte
	binary.Read(r, binary.LittleEndian, &order)
	byteOrder := WkbByte2Order(order)
	var wkbType uint32
	binary.Read(r, byteOrder, &wkbType)
	binary.Read(r, byteOrder, &this.Point2D)
	return true
}

func (this *GeoPoint) To(mode GeoMode) []byte {
	switch mode {
	case WKB:
		return this.toWKB()
	case WKT:
		// todo
	case GAIA:
		return this.toGAIA()
	}
	return nil
}

func (this *GeoPoint) toWKB() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, byte(WkbLittle))
	binary.Write(&buf, binary.LittleEndian, uint32(TGeoPoint))
	binary.Write(&buf, binary.LittleEndian, this.Point2D)
	return buf.Bytes()
}

func (this *GeoPoint) toGAIA() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, byte(0))

	var info GAIAInfo
	var bbox base.Rect2D
	bbox.Max = this.Point2D
	bbox.Min = this.Point2D
	info.Init(bbox, 0)

	binary.Write(&buf, binary.LittleEndian, info.To())
	binary.Write(&buf, binary.LittleEndian, int32(1)) // type
	binary.Write(&buf, binary.LittleEndian, this.Point2D)
	binary.Write(&buf, binary.LittleEndian, byte(0XFE))
	return buf.Bytes()
}

// ========================================================= //

func (this *GeoPoint) IsRelate(mode base.SpatialMode, rect base.Rect2D) bool {
	switch mode {
	// 以下a/b可调换
	case base.Intersects, base.Undefined, base.BBoxIntersects:
		return rect.IsIntersectsPnt(this.Point2D)
	case base.Disjoint:
		return rect.IsDisjointPnt(this.Point2D)
	case base.Touches:
		return rect.IsTouchesPnt(this.Point2D)
	// 以下不能换顺序
	case base.Within:
		// bbox在 geo的内部（且边界不接触）
		return rect.IsContainsPnt(this.Point2D)
	case base.CoveredBy:
		// bbox在 geo的内部（且边界可接触）
		return rect.IsCoversPnt(this.Point2D)
	// 这几个不适用
	case base.Crosses, base.Overlaps, base.Equals, base.Contains, base.Covers:
		return false
	default:
		var im base.D9IM
		im.Init(string(mode))
		imRes := this.CalcRelateIM(rect)
		return imRes.MatchIM(im)
		// return this.isRelatedRect(im, rect)
	}
	// return false
}

// 计算点和rect的九交模型矩阵；
func (this *GeoPoint) CalcRelateIM(rect base.Rect2D) (out base.D9IM) {
	// 先初始化为 *
	out.Init("*********")

	// 点只需要关心在rect内，在边界，在外面 三种关系
	if rect.IsContainsPnt(this.Point2D) {
		out.Set(base.I, base.I, '0')
	} else if rect.IsTouchesPnt(this.Point2D) {
		out.Set(base.I, base.B, '0')
	} else { // 剩下只有在外的了
		out.Set(base.I, base.E, '0')
	}
	return
}

// ================================================================ //
/*
func (this *GeoPoint) IsEquals(geo Geometry) bool {
	if geoPnt, ok := geo.(*GeoPoint); ok {
		return this.Point2D == geoPnt.Point2D
	}
	return false
}

func (this *GeoPoint) IsRelated(mode base.SpatialMode, geo Geometry) bool {
	switch mode {
	case base.Equals:
		return this.IsEquals(geo)
	case base.BBoxIntersects:
		return this.GetBounds().IsIntersects(geo.GetBounds())
	case base.Intersects, base.Undefined:
		return this.IsIntersects(geo)
	case base.Disjoint:
		return this.IsDisjoint(geo)
	case base.Touches:
		return this.IsTouches(geo)
	case base.Within:
		return this.IsWithin(geo)
	case base.CoveredBy:
		return this.IsCoveredBy(geo)
	case base.Overlaps:
		return this.IsOverlaps(geo)
	case base.Contains:
		return this.IsContains(geo)
	case base.Covers:
		return this.IsCovers(geo)
	case base.Crosses: // 点无法cross任何对象
		return false
	default:
		var im base.D9IM
		im.Init(string(mode))
		imRes := this.CalcRelateIM(geo)
		return imRes.MatchIM(im)
	}
}

func (this *GeoPoint) CalcRelateIM(geo Geometry) (im base.D9IM) {
	im.Init("*********")
	if pnt, ok := geo.(*GeoPoint); ok {
		if this.IsEquals(pnt) {
			im.Set(base.I, base.I, '0')
		}
	} else {
		im = geo.CalcRelateIM(this)
		im.Invert()
	}
	return
}

func (this *GeoPoint) IsCovers(geo Geometry) bool {
	// 只有点和点才能overlap
	if pnt, ok := geo.(*GeoPoint); ok {
		return this.IsEquals(pnt)
	}
	return false
}

func (this *GeoPoint) IsContains(geo Geometry) bool {
	// 只有点和点才能overlap
	if pnt, ok := geo.(*GeoPoint); ok {
		return this.IsEquals(pnt)
	}
	return false
}

func (this *GeoPoint) IsOverlaps(geo Geometry) bool {
	// 只有点和点才能overlap
	// todo 未来应支持其它0维对象
	if pnt, ok := geo.(*GeoPoint); ok {
		return this.IsEquals(pnt)
	}
	return false
}

func (this *GeoPoint) IsCoveredBy(geo Geometry) bool {
	if _, ok := geo.(*GeoPoint); ok {
		return this.IsEquals(geo)
	}
	return geo.IsRelated(base.Covers, this)
}

func (this *GeoPoint) IsWithin(geo Geometry) bool {
	if _, ok := geo.(*GeoPoint); ok {
		return this.IsEquals(geo)
	}
	return geo.IsRelated(base.Contains, this)
}

func (this *GeoPoint) IsTouches(geo Geometry) bool {
	if _, ok := geo.(*GeoPoint); ok {
		return false // 点和点 只有内部，不适用
	}
	return geo.IsRelated(base.Touches, this)
}

func (this *GeoPoint) IsDisjoint(geo Geometry) bool {
	if pnt, ok := geo.(*GeoPoint); ok {
		return !this.IsEquals(pnt)
	}
	return geo.IsRelated(base.Disjoint, this)
}

// 内部或边界有交集即可
func (this *GeoPoint) IsIntersects(geo Geometry) bool {
	if pnt, ok := geo.(*GeoPoint); ok {
		return this.IsEquals(pnt)
	}
	return geo.IsRelated(base.Intersects, this)
}
*/
