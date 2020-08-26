package geometry

import "gogis/base"

// 几何对象类型定义
type GeoType int

const (
	// _iota
	TGeoEmpty      = iota // 空对象
	TGeoPoint             // 点
	TGeoPolyline          // 线
	TGeoPolygon           // 面
	TGeoMultiPoint        // 多点
	// TGeoMultiLineString        // 多线
	// TGeoMultiPolygon           // 多面
	TGeoCollection // 集合
	// Point              GeoType = 0 // 点
	// LineString         GeoType = 1 // 线
	// Polygon            GeoType = 2 // 面
	// MultiPoint         GeoType = 3 // 多点
	// MultiLineString    GeoType = 4 // 多线
	// MultiPolygon       GeoType = 5 // 多面
	// GeometryCollection GeoType = 6 // 集合
)

// 几何对象外部存储方式定义
type GeoMode int

const (
	WKB     GeoMode = 0
	WKT     GeoMode = 1
	GeoJson GeoMode = 2
)

type Geometry interface {
	Type() GeoType
	GetBounds() base.Rect2D
	// ID() int64
	// Draw() // todo 绘制参数
	// From(data []byte, mode GeoMode) bool
	// To(mode GeoMode) []byte
}

type Point2D struct {
	X float64
	Y float64
}
