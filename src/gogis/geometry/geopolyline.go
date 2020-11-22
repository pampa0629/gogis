package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
	"gogis/draw"
)

type GeoPolyline struct {
	Points [][]base.Point2D
	BBox   base.Rect2D
}

func (this *GeoPolyline) Type() GeoType {
	// 	if len(this.Points) == 1 {
	// 		return TGeoLineString
	// 	} else if len(this.Points) > 1 {
	// 		return TGeoMultiLineString
	// 	}
	return TGeoPolygon
}

func (this *GeoPolyline) GetBounds() base.Rect2D {
	return this.BBox
}

func (this *GeoPolyline) ComputeBounds() base.Rect2D {
	this.BBox.Init()
	for _, points := range this.Points {
		this.BBox.Union(base.ComputeBounds(points))
	}
	return this.BBox
}

func (this *GeoPolyline) Draw(canvas *draw.Canvas) {
	var line = new(draw.Polyline)
	line.Points = make([][]draw.Point, len(this.Points))
	// line := make([][]draw.Point, len(this.Points))
	for i, v := range this.Points {
		line.Points[i] = canvas.Params.Forwards(v)
	}
	canvas.DrawPolyPolyline(line)
	// canvas.Stroke()
}

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

	switch mode {
	case WKB:
		if len(data) >= 40 {
			buf := bytes.NewBuffer(data)
			var order byte
			binary.Read(buf, binary.LittleEndian, &order)
			byteOrder := WkbByte2Order(order)
			var wkbType WkbGeoType
			binary.Read(buf, byteOrder, &wkbType)
			if wkbType == WkbLineString { // 简单线
				this.Points = append(this.Points, Bytes2WkbLinearRing(byteOrder, buf))
			} else if wkbType == WkbMultiLineString { // 复杂线
				var numLineStrings uint32
				binary.Read(buf, byteOrder, &numLineStrings)
				this.Points = make([][]base.Point2D, numLineStrings)
				for i := uint32(0); i < numLineStrings; i++ {
					var order byte
					binary.Read(buf, binary.LittleEndian, &order)
					byteOrder := WkbByte2Order(order)
					var wkbType WkbGeoType
					binary.Read(buf, byteOrder, &wkbType)
					this.Points[i] = Bytes2WkbLinearRing(byteOrder, buf)
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

func (this *GeoPolyline) To(mode GeoMode) []byte {
	switch mode {
	case WKB:
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
	case WKT:
		// todo
	}
	return nil
}

// ID() int64
