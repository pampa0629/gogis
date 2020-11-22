package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
)

// 几何对象类型定义
type WkbGeoType int32

const (
	WkbPoint           WkbGeoType = 1 // 点
	WkbLineString      WkbGeoType = 2 // 线
	WkbPolygon         WkbGeoType = 3 // 面
	WkbMultiPoint      WkbGeoType = 4 // 多点
	WkbMultiLineString WkbGeoType = 5 // 多线
	WkbMultiPolygon    WkbGeoType = 6 // 多面
	// WkbGeometryCollection WkbGeoType = 7 // 集合
)

// enum  WKBGeometryType {
// 	wkbPoint  = 1,
// 	wkbLineString  = 2,
// 	wkbPolygon  = 3,
// 	wkbTriangle  = 17
// 	wkbMultiPoint  = 4,
// 	wkbMultiLineString  = 5,
// 	wkbMultiPolygon  = 6,
// 	wkbGeometryCollection  = 7,
// 	wkbPolyhedralSurface  = 15,
// 	wkbTIN  = 16
// 	wkbPointZ  = 1001,
// 	wkbLineStringZ  = 1002,
// 	wkbPolygonZ  = 1003,
// 	wkbTrianglez = 1017
// 	wkbMultiPointZ  = 1004,
// 	wkbMultiLineStringZ  = 1005,
// 	wkbMultiPolygonZ  = 1006,
// 	wkbGeometryCollectionZ  = 1007,
// 	wkbPolyhedralSurfaceZ  = 1015,
// 	wkbTINZ  = 1016
// 	wkbPointM  = 2001,
// 	wkbLineStringM  = 2002,
// 	wkbPolygonM  = 2003,
// 	wkbTriangleM  = 2017
// 	wkbMultiPointM  = 2004,
// 	wkbMultiLineStringM  = 2005,
// 	wkbMultiPolygonM  = 2006,
// 	wkbGeometryCollectionM  = 2007,
// 	wkbPolyhedralSurfaceM  = 2015,
// 	wkbTINM  = 2016
// 	wkbPointZM  = 3001,
// 	wkbLineStringZM  = 3002,
// 	wkbPolygonZM  = 3003,
// 	wkbTriangleZM  = 3017
// 	wkbMultiPointZM  = 3004,
// 	wkbMultiLineStringZM  = 3005,
// 	wkbMultiPolygonZM  = 3006,
// 	wkbGeometryCollectionZM  = 3007,
// 	wkbPolyhedralSurfaceZM  = 3015,
// 	wkbTinZM  = 3016,

// 字节顺序，一个字节存储
type WkbByteOrder byte

const (
	WkbBig    WkbByteOrder = 0 // 大尾端
	WkbLittle WkbByteOrder = 1 // 小尾端，默认都用小端
)

// 读取一个字节，确定 wkb后续内容的字节顺序
func WkbByte2Order(wkbByte byte) (byteOrder binary.ByteOrder) {
	if wkbByte == byte(WkbBig) {
		byteOrder = binary.BigEndian
	} else if wkbByte == byte(WkbLittle) {
		byteOrder = binary.LittleEndian
	} else {
		panic("")
	}
	return
}

// LinearRing{
// 	uint32  numPoints;
// 	Point  points[numPoints]}
// 从字节缓存中，读取 wkb 简单线；只包括 点个数和坐标数组
func WkbLinearRing2Bytes(points []base.Point2D, buf *bytes.Buffer) {
	binary.Write(buf, binary.LittleEndian, uint32(len(points)))
	binary.Write(buf, binary.LittleEndian, points)
}

func Bytes2WkbLinearRing(byteOrder binary.ByteOrder, buf *bytes.Buffer) (points []base.Point2D) {
	var numPoints uint32
	binary.Read(buf, byteOrder, &numPoints)
	points = make([]base.Point2D, numPoints)
	binary.Read(buf, byteOrder, points)
	return
}

// WKBLineString
// byte  byteOrder;                               //字节序
// static  uint32  wkbType = 2;                    //几何体类型
// uint32  numPoints;                                  //点的个数
// Point  points[numPoints]}                 //点的坐标数组
// 把Wkb简单线写入字节缓存
func WkbLineString2Bytes(points []base.Point2D, buf *bytes.Buffer) {
	binary.Write(buf, binary.LittleEndian, byte(WkbLittle))
	binary.Write(buf, binary.LittleEndian, uint32(WkbLineString))
	WkbLinearRing2Bytes(points, buf)
}

// WKBPolygon{
// byte  byteOrder;                               //字节序
// static  uint32  wkbType= 3;                    //几何体类型
// uint32  numRings;                                   //线串的个数
// LinearRing rings[numRings]}           //线串（环）的数组

// 存储简单面到 wkb bytes中
func WkbPolygon2Bytes(points [][]base.Point2D, buf *bytes.Buffer) {
	binary.Write(buf, binary.LittleEndian, byte(WkbLittle))
	binary.Write(buf, binary.LittleEndian, uint32(WkbPolygon))
	binary.Write(buf, binary.LittleEndian, uint32(len(points)))
	for _, v := range points {
		WkbLinearRing2Bytes(v, buf)
	}
}

// 从字节缓存中，读取 wkb 简单面；只包括 点个数和坐标数组，不包括字节序和几何类型
func Bytes2WkbPolygon(byteOrder binary.ByteOrder, buf *bytes.Buffer) (points [][]base.Point2D) {
	var numRings uint32
	binary.Read(buf, byteOrder, &numRings)
	points = make([][]base.Point2D, numRings)
	for i := uint32(0); i < numRings; i++ {
		points[i] = Bytes2WkbLinearRing(byteOrder, buf)
	}
	return
}
