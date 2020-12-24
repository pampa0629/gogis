package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
	"gogis/draw"
)

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

func (this *GeoPoint) Draw(canvas *draw.Canvas) {
	pnt := canvas.Params.Forward(this.Point2D)
	canvas.DrawPoint(pnt)
}

// wkb:
// byte  byteOrder;                        //字节序 1
// static  uint32  wkbType= 1;             //几何体类型 4
// Point  point}                             //点的坐标 8*2

func (this *GeoPoint) From(data []byte, mode GeoMode) (result bool) {
	switch mode {
	case WKB:
		result = this.fromWKB(data)
	case WKT:
		// todo
	case GAIA:
		result = this.fromGAIA(data)
	}
	return
}

func (this *GeoPoint) fromGAIA(data []byte) bool {
	// fmt.Println("data:", data)
	if len(data) >= 60 {
		buf := bytes.NewBuffer(data)
		var begin byte
		binary.Read(buf, binary.LittleEndian, &begin)
		if begin == byte(0X00) {
			var info GAIAInfo
			byteOrder := info.From(buf)
			var geoType int32
			binary.Read(buf, byteOrder, &geoType)
			if geoType == 1 {
				binary.Read(buf, byteOrder, &this.Point2D)
				var end byte
				binary.Read(buf, binary.LittleEndian, &end)
				return true
			} else if geoType == 4 {
				// todo 暂时先这么处理，后续增加多点对象
				var count int32
				binary.Read(buf, binary.LittleEndian, &count)
				for i := int32(0); i < count; i++ {
					var mark byte
					binary.Read(buf, binary.LittleEndian, &mark)
					var geoType int32
					binary.Read(buf, byteOrder, &geoType)
					binary.Read(buf, byteOrder, &this.Point2D)
				}
			}
			return true
		}
	}
	return false
}

func (this *GeoPoint) fromWKB(data []byte) bool {
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
	return false
}

func (this *GeoPoint) To(mode GeoMode) []byte {
	switch mode {
	case WKB:
		return this.ToWKB()
	case WKT:
		// todo
	}
	return nil
}

func (this *GeoPoint) ToWKB() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, byte(WkbLittle))
	binary.Write(&buf, binary.LittleEndian, uint32(TGeoPoint))
	binary.Write(&buf, binary.LittleEndian, this.Point2D)
	return buf.Bytes()
}
