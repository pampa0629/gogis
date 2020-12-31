package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
	"gogis/draw"
	"io"
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
