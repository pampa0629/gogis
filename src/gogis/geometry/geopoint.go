package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
	"gogis/draw"
)

type GeoPoint struct {
	base.Point2D
}

func (this *GeoPoint) Type() GeoType {
	return TGeoPoint
}

func (this *GeoPoint) GetBounds() base.Rect2D {
	return base.NewRect2D(this.X, this.Y, this.X, this.Y)
}

func (this *GeoPoint) Draw(canvas *draw.Canvas) {
	pnt := canvas.Params.Forward(this.Point2D)
	canvas.DrawPoint(pnt)
}

// wkb:
// byte  byteOrder;                        //字节序 1
// static  uint32  wkbType= 1;             //几何体类型 4
// Point  point}                             //点的坐标 8*2

func (this *GeoPoint) From(data []byte, mode GeoMode) bool {
	switch mode {
	case WKB:
		if len(data) >= 21 {
			buf := bytes.NewBuffer(data)
			var order byte
			binary.Read(buf, binary.LittleEndian, &order)
			byteOrder := WkbByte2Order(order)
			var wkbType uint32
			binary.Read(buf, byteOrder, &wkbType)
			binary.Read(buf, byteOrder, &this.Point2D)
			return true
		}
	case WKT:
		// todo
	}
	return false
}

func (this *GeoPoint) To(mode GeoMode) []byte {
	switch mode {
	case WKB:
		var buf bytes.Buffer

		// 数字转 []byte
		binary.Write(&buf, binary.LittleEndian, byte(WkbLittle))
		binary.Write(&buf, binary.LittleEndian, uint32(TGeoPoint))
		binary.Write(&buf, binary.LittleEndian, this.Point2D)
		return buf.Bytes()
	case WKT:
		// todo
	}
	return nil
}

// ID() int64
