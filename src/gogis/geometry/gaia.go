package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
	"io"
)

type GAIAInfo struct {
	// static byte byteOrdering = 1; //字节序：小端存储
	// int32 srid; //坐标系 ID
	// Rect mbr; //对象的坐标范围
	// static byte gaiaMBR=0x7c; //MBR 结束标识
	order byte  // 字节顺序
	srid  int32 // 坐标系id
	bbox  base.Rect2D
	mark  byte
}

func (this *GAIAInfo) Init(bbox base.Rect2D, srid int32) {
	this.bbox = bbox
	this.srid = srid
	this.order = 1
	this.mark = 0x7c
}

func (this *GAIAInfo) From(r io.Reader) binary.ByteOrder {
	binary.Read(r, binary.LittleEndian, &this.order)
	byteOrder := GAIAByte2Order(this.order)
	binary.Read(r, byteOrder, &this.srid)
	binary.Read(r, byteOrder, &this.bbox)
	binary.Read(r, byteOrder, &this.mark)
	if this.mark != byte(0X7C) {
		panic("gaia info mark error:" + string(this.mark))
	}
	return byteOrder
}

func (this *GAIAInfo) To() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, this.order)
	binary.Write(&buf, binary.LittleEndian, this.srid)
	binary.Write(&buf, binary.LittleEndian, this.bbox)
	binary.Write(&buf, binary.LittleEndian, this.mark)
	return buf.Bytes()
}

// 字节顺序，一个字节存储
type GAIAByteOrder byte

const (
	GaiaBig    GAIAByteOrder = 0 // 大尾端
	GaiaLittle GAIAByteOrder = 1 // 小尾端，默认都用小端
)

// 确定后续内容的字节顺序
func GAIAByte2Order(gaiaByte byte) (byteOrder binary.ByteOrder) {
	switch gaiaByte {
	case byte(GaiaBig):
		byteOrder = binary.BigEndian
	case byte(GaiaLittle):
		byteOrder = binary.LittleEndian
	}
	return
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

// LineStringEntity {
// 	static byte gaiaEntityMark = 0x69;//子对象标识
// 	static int32 geoType = 2; //Geometry 类型标识
// 	int32 numPoints; //点个数
// 	Point[] pnts[numPoints]; //每个点的坐标值
// 	}
func gaia2LineString(r io.Reader, byteOrder binary.ByteOrder) []base.Point2D {
	var mark byte
	binary.Read(r, binary.LittleEndian, &mark)
	if mark == 0X69 {
		var geoType int32
		binary.Read(r, byteOrder, &geoType)
		if geoType == 2 {
			return gaia2Ring(r, byteOrder)
		}
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
