package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
	"gogis/draw"
	"io"
	"math"
)

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

func (this *GeoPolygon) ComputeBounds() base.Rect2D {
	this.BBox.Init()
	for _, polygon := range this.Points {
		// 只有第一个是岛，需要计算bounds；后面有也是洞，可以不理会
		if len(polygon) >= 1 {
			this.BBox.Union(base.ComputeBounds(polygon[0]))
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
				this.Points[i][ii] = prjc.Do(vv)
			}
		}
	}
}

// 抽稀一个封闭区域
func thinOneGon(points []base.Point2D, dis2 float64) (newPnts []base.Point2D) {
	newPnts = thinOneLine(points, dis2)
	// 不够三个点，就不要了
	count := len(newPnts)
	if count < 3 {
		// return newPnts[0:1]
		return nil
	}
	// 封闭区域
	if newPnts[0] != newPnts[count-1] {
		newPnts = append(newPnts, newPnts[0])
	}
	return
}

func (this *GeoPolygon) Thin(dis2 float64) Geometry {
	var newgeo GeoPolygon
	newgeo.BBox = this.BBox
	newgeo.id = this.id
	newgeo.Points = make([][][]base.Point2D, 0, len(this.Points))
	for i, _ := range this.Points {
		pntss := make([][]base.Point2D, 0, len(this.Points[i]))
		for _, v := range this.Points[i] {
			pnts := thinOneGon(v, dis2)
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
	copy(this.Points[0][0][0:4], bbox.ToPoints())
	this.Points[0][0][4] = this.Points[0][0][0]
}

// 自己的bbox（含子对象的），被bbox包裹
func (this *GeoPolygon) BboxesIsCovered(bbox base.Rect2D) bool {
	if bbox.IsCover(this.BBox) {
		return true
	}
	// 如果有多个子对象，则每个都要单独判断；有一个就ok
	if len(this.Points) > 1 {
		for _, v := range this.Points {
			subBbox := base.ComputeBounds(v[0])
			if bbox.IsCover(subBbox) {
				return true
			}
		}
	}
	return false
}

// 子对象的bbox，是否与bbox相交；有一个相交就返回true
// 没有子对象，就用自己的
func (this *GeoPolygon) SubBboxesIsIntersect(bbox base.Rect2D) bool {
	if len(this.Points) > 1 {
		for _, v := range this.Points {
			subBbox := base.ComputeBounds(v[0])
			if subBbox.IsIntersect(bbox) {
				return true
			}
		}
	} else {
		if this.BBox.IsIntersect(bbox) {
			return true
		}
	}
	return false
}

// 判断是否和bbox相交（或者被bbox所包裹）
func (this *GeoPolygon) IsIntersect(bbox base.Rect2D) bool {
	// 先用bbox是否conver进行判断
	if this.BboxesIsCovered(bbox) {
		return true
	}
	// 子对象的bbox都没有一个相交的，肯定不会相交了
	if !this.SubBboxesIsIntersect(bbox) {
		return false
	}

	// 如果bbox有任意一个顶点在一个子对象的外环之内，且bbox整体不在该子对象的某个内环之内，返回true
	pnts := this.BBox.ToPoints()
	for _, v := range this.Points {
		oneInPolygon := false
		for _, pnt := range pnts {
			if PntIsWithin(pnt, v[0]) {
				oneInPolygon = true
				break
			}
		}
		if oneInPolygon {
			for j := 1; j < len(v); j++ {
				if PntsIsWithin(pnts, v[j]) {
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

// 判断几个点是否都在一个环之内
func PntsIsWithin(pnts []base.Point2D, ring []base.Point2D) bool {
	for _, v := range pnts {
		if !PntIsWithin(v, ring) {
			return false
		}
	}
	return true
}

// 判断点是否在一个环之内
// 采用射线法，看水平方向射线与所有边的交点个数的奇偶性；奇数为内，偶数为外
// 点若在线上，不属于在环内
func PntIsWithin(pnt base.Point2D, ring []base.Point2D) bool {
	count := 0
	for i := 1; i < len(ring); i++ {
		p1 := ring[i-1]
		p2 := ring[i]
		if PntOnSegment(pnt, p1, p2) {
			return false
		}
		count += PntRaySegment(pnt, p1, p2)
	}

	if count%2 == 0 {
		return false
	}
	return true
}

// 点往X+做射线，看是否与线段有交点；有返回1，没有返回0
func PntRaySegment(p, p1, p2 base.Point2D) int {
	// 水平的线段，认为没有交点（交点可能有无限多个）
	if p1.Y == p2.Y {
		return 0
	}
	// 如果点的射线，与一个节点相交，则忽略y值较小的情况
	miny := math.Min(p1.Y, p2.Y)
	maxy := math.Max(p1.Y, p2.Y)
	if p.Y == p1.Y || p.Y == p2.Y {
		if p.Y == miny {
			return 0
		} else {
			return 1
		}
	}
	// 如果点的Y 在 p1.y 和 p2.y 之间，则说明有交点
	if p.Y > miny && p.Y < maxy {
		return 1
	}
	return 0
}

// 判断点是否在线上
func PntOnSegment(p, p1, p2 base.Point2D) bool {
	// 这里先假定两个节点不能相同
	if p1 == p2 {
		panic("two point of a segment cannot be same.")
	}
	// 等于某个节点，自然在了
	if p == p1 || p == p2 {
		return true
	}
	// 点必须在两个点的范围内
	if p.X >= math.Min(p1.X, p2.X) && p.X <= math.Max(p1.X, p2.X) &&
		p.Y >= math.Min(p1.Y, p2.Y) && p.Y <= math.Max(p1.Y, p2.Y) {
		if base.IsEqual(CrossX(p1, p, p2), 0) {
			return true
		}
	}
	return false
}

// 叉乘；若结果为0，说明共线
func CrossX(a, b, c base.Point2D) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (c.X-a.X)*(b.Y-a.Y)
}

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
