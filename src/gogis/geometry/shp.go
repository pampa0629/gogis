package geometry

import (
	"encoding/binary"
	"gogis/base"
	"io"
)

type ShpType int32

const (
	ShpNull     = 0
	ShpPoint    = 1
	ShpPolyLine = 3
	ShpPolygon  = 5
	// todo
	ShpMultiPoint  = 8
	ShpPointZ      = 11
	ShpPolyLineZ   = 13
	ShpPolygonZ    = 15
	ShpMultiPointZ = 18
	ShpPointM      = 21
	ShpPolyLineM   = 23
	ShpPolygonM    = 25
	ShpMultiPointM = 28
	ShpMultiPatch  = 31
)

// 类型之间对应
func ShpType2Geo(shptype ShpType) (geoType GeoType) {
	switch shptype {
	case ShpPoint:
		geoType = TGeoPoint
	case ShpPolyLine:
		geoType = TGeoPolyline
	case ShpPolygon:
		geoType = TGeoPolygon
		// todo more
	}
	return
}

// 读取 polyline和polygon共同的部分
func loadShpOnePolyHeader(r io.Reader) (bbox base.Rect2D, numParts, numPoints int32) {
	var shptype int32
	binary.Read(r, binary.LittleEndian, &shptype)
	// 这里合并处理
	binary.Read(r, binary.LittleEndian, &bbox)
	binary.Read(r, binary.LittleEndian, &numParts)
	binary.Read(r, binary.LittleEndian, &numPoints)
	return
}
