package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
	"gogis/draw"
	"io"
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

func (this *GeoPoint) Clone() Geometry {
	res := new(GeoPoint)
	res.id = this.id
	res.Point2D = this.Point2D
	return res
}

func (this *GeoPoint) Draw(canvas *draw.Canvas) {
	pnt := canvas.Params.Forward(this.Point2D)
	canvas.DrawPoint(pnt)
}

func (this *GeoPoint) ConvertPrj(prjc *base.PrjConvert) {
	if prjc != nil {
		this.Point2D = prjc.DoOne(this.Point2D)
	}
}

// 点就只能返回自身
func (this *GeoPoint) Thin(dis2 float64) Geometry {
	return this
}

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
