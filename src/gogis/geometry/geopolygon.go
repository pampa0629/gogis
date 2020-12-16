package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
	"gogis/draw"
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

func (this *GeoPolygon) From(data []byte, mode GeoMode) bool {
	this.Points = this.Points[0:0] // 清空

	switch mode {
	case WKB:
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
			this.ComputeBounds()
			return true
		}
	case WKT:
		// todo
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
