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

	switch mode {
	case WKB:
		result = this.fromWKB(data)
		this.ComputeBounds()
	// todo
	// case WKT:
	case GAIA:
		result = this.fromGAIA(data)
	}

	return
}

func (this *GeoPolygon) fromGAIA(data []byte) bool {
	// fmt.Println("data:", data)
	if len(data) >= 60 {
		buf := bytes.NewBuffer(data)
		var begin byte
		binary.Read(buf, binary.LittleEndian, &begin)
		if begin == byte(0X00) {

			var info GAIAInfo
			byteOrder := info.From(buf)
			this.BBox = info.bbox

			var geoType int32
			binary.Read(buf, byteOrder, &geoType)
			if geoType == 6 {
				// int32 numPolygon; //子对象个数
				// PolygonEntity[] polygons[numPolygon]; //子对象数据
				var subCount int32
				binary.Read(buf, byteOrder, &subCount)
				this.Points = make([][][]base.Point2D, subCount)
				for i := int32(0); i < subCount; i++ {
					// PolygonEntity {
					// 	static byte gaiaEntityMark = 0x69; //子对象标识
					// 	PolygonData data; //子对象数据
					// 	}
					var mark byte
					binary.Read(buf, byteOrder, &mark)
					if mark == 0x69 {
						this.Points[i] = gaia2Polygon(buf, byteOrder)
					}
				}
				return true
			}
		}
	}
	return false
}

// PolygonData {
// 	static int32 geoType = 3; //Geometry 类型标识
// 	int32 numInteriors; //环的总个数
// 	Ring exteriorRing; //外环对象
// 	Ring[] interiorRings[numInteriors]; //内环对象
// 	}

func gaia2Polygon(r io.Reader, byteOrder binary.ByteOrder) [][]base.Point2D {
	var geoType int32
	binary.Read(r, byteOrder, &geoType)
	if geoType == 3 {
		var count int32
		binary.Read(r, byteOrder, &count)
		pnts := make([][]base.Point2D, count)
		// pnts[0] = gaia2Ring(r, byteOrder)
		for i := 0; i < int(count); i++ {
			pnts[i] = gaia2Ring(r, byteOrder)
		}
		return pnts
	}
	return nil
}

// Ring { //由点组成的环形
// 	int32 numPoints; //点个数
// 	Point[] pnts[numPoints]; //点坐标
// 	}
func gaia2Ring(r io.Reader, byteOrder binary.ByteOrder) []base.Point2D {
	var pntCount int32
	binary.Read(r, byteOrder, &pntCount)
	pnts := make([]base.Point2D, pntCount)
	binary.Read(r, byteOrder, pnts)
	return pnts
}

func (this *GeoPolygon) fromWKB(data []byte) bool {
	if len(data) >= 40 {
		buf := bytes.NewBuffer(data)
		var order byte
		binary.Read(buf, binary.LittleEndian, &order)
		byteOrder := WkbByte2Order(order)
		var wkbType WkbGeoType
		binary.Read(buf, byteOrder, &wkbType)
		if wkbType == WkbPolygon { // 简单面
			this.Points = append(this.Points, Bytes2WkbPolygon(byteOrder, buf))
		} else if wkbType == WkbMultiPolygon { // 复杂面
			var numPolygons uint32
			binary.Read(buf, byteOrder, &numPolygons)
			this.Points = make([][][]base.Point2D, numPolygons)
			for i := uint32(0); i < numPolygons; i++ {
				var order byte
				binary.Read(buf, binary.LittleEndian, &order)
				byteOrder := WkbByte2Order(order)
				var wkbType WkbGeoType
				binary.Read(buf, byteOrder, &wkbType)
				this.Points[i] = Bytes2WkbPolygon(byteOrder, buf)
			}
		}
		return true
	}
	return false
}

func (this *GeoPolygon) To(mode GeoMode) []byte {
	switch mode {
	case WKB:
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
	case WKT:
		// todo
	}
	return nil
}

// ID() int64
