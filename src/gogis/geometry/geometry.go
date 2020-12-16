package geometry

import (
	"gogis/base"
)

// 几何对象类型定义
type GeoType int

const (
	// _iota
	TGeoEmpty      GeoType = 0  // 空对象
	TGeoPoint      GeoType = 1  // 点
	TGeoPolyline   GeoType = 5  // 线2，等同于多线5
	TGeoPolygon    GeoType = 3  // 面
	TGeoMultiPoint GeoType = 11 // 多点
	// TGeoMultiLineString        // 多线
	// TGeoMultiPolygon           // 多面
	TGeoCollection GeoType = 100 // 集合
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
	// Shape   GeoMode = 3 // shape格式的二进制存储
)

type Geometry interface {
	Type() GeoType
	GetBounds() base.Rect2D

	// 加载和保存为特定格式
	From(data []byte, mode GeoMode) bool
	To(mode GeoMode) []byte

	// ID
	SetID(id int64)
	GetID() int64

	// todo 未来也作为单独的接口来对待
	// HitTest() bool
}

// id 的处理单独用一个结构来搞定，避免重复代码
type GeoID struct {
	id int64
}

// ID
func (this *GeoID) SetID(id int64) {
	this.id = id
}

func (this *GeoID) GetID() int64 {
	return this.id
}
