package geometry

import (
	"gogis/algorithm"
	"gogis/base"
)

// 几何对象类型定义
type GeoType int

const (
	// _iota
	TGeoEmpty    GeoType = 0 // 空对象
	TGeoPoint    GeoType = 1 // 点
	TGeoPolyline GeoType = 5 // 线2，等同于多线5
	TGeoPolygon  GeoType = 3 // 面
	// TGeoMultiPoint GeoType = 11 // 多点
	// TGeoMultiLineString        // 多线
	// TGeoMultiPolygon           // 多面
	// TGeoCollection GeoType = 100 // 集合

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
	GAIA    GeoMode = 3 // spatialite 的二进制格式
	Shape   GeoMode = 4 // shape格式的二进制存储
)

// func CloneGeo(geo Geometry) (newgeo Geometry) {
// 	if geo != nil {
// 		data := geo.To(WKB)
// 		newgeo = CreateGeo(geo.Type())
// 		newgeo.From(data, WKB)
// 	}
// 	return
// }

type Geometry interface {
	Type() GeoType
	// 得到外接矩形
	GetBounds() base.Rect2D
	// 得到几何对象的空间维度，点为0，线为1，面为2
	Dim() int
	// 得到对象边界的维度，没有为-1
	DimB() int
	// 得到组成内部或边界的点串
	// GetPnts(ibe base.IBE) [][]base.Point2D
	// 得到子对象个数
	SubCount() int

	// 对自身做投影转化
	ConvertPrj(prjc *base.PrjConvert)
	// 克隆，返回一模一样的
	Clone() Geometry

	// 加载和保存为特定格式
	From(data []byte, mode GeoMode) bool
	To(mode GeoMode) []byte

	// 抽稀，作为金字塔使用；返回抽稀后的对象，若抽稀后不成立，则返回nil
	// dis2 为距离的平方；当两个点的距离平方小于dis2时, 该点被忽略
	// angle: 角度(非弧度);  当拐角大于angle时,该点被忽略
	Thin(dis2, angle float64) Geometry

	// 和rect之间的空间关系是否满足mode要求；用于和rect的拉框查询等
	IsRelate(mode base.SpatialMode, rect base.Rect2D) bool
	// 与另一个geometry之间的空间判断函数：通过 GeoRelation实现
	// IsRelate(mode base.SpatialMode, geo Geometry) bool

	// ID
	SetID(id int64)
	GetID() int64

	// todo 未来也作为单独的接口来对待
	// HitTest() bool
}

// 0维对象必须实现以下函数
type Geo0 interface {
	GetPnts() []base.Point2D
}

// 1维对象必须实现以下函数
type Geo1 interface {
	// 得到边界点
	GetPntsB() []base.Point2D
	GetSubLine(num int) algorithm.Line
}

// 2维对象必须实现以下函数
type Geo2 interface {
	GetSubRegion(num int) algorithm.Region
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

type CreateGeoFunc func() Geometry

var gCreateGeos map[GeoType]CreateGeoFunc

// 支持用户自定义geometry类型
func RegisterGeo(geotype GeoType, create CreateGeoFunc) {
	if gCreateGeos == nil {
		gCreateGeos = make(map[GeoType]CreateGeoFunc)
	}
	gCreateGeos[geotype] = create
}

func CreateGeo(geoType GeoType) Geometry {
	creatFunc, ok := gCreateGeos[geoType]
	if ok {
		return creatFunc()
	}
	return nil
}

// func CreateGeo(geoType GeoType) Geometry {
// 	switch geoType {
// 	case TGeoEmpty:
// 		return nil
// 	case TGeoPoint:
// 		return &GeoPoint{}
// 	case TGeoPolyline:
// 		return &GeoPolyline{}
// 	case TGeoPolygon:
// 		return &GeoPolygon{}
// 	case TGeoMultiPoint:
// 	case TGeoCollection:
// 		return nil
// 	}
// 	return nil
// }
