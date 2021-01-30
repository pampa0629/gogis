package data

import (
	"gogis/base"
	"gogis/geometry"
	"gogis/index"
)

// 属性查询条件定义
type QueryDef struct {
	Fields      []string         // 需要哪些字段，如果为空，则查询所有字段
	Where       string           // sql中的where语句，支持 and/or/()，如：Field1="abc" and (Field2>=10 or Field2<=0) ...
	SpatialMode base.SpatialMode // 空间查询模式
	// todo 必须支持 GetBounds() 接口，包括 Rect2D、Geometry和 Featureset 等
	SpatialObj interface{} // 查询对象，即用本对象去查询符合条件的对象
}

// 空间查询类
type SpatailQuery struct {
	obj  interface{}
	mode base.SpatialMode
}

func (this *SpatailQuery) Init(obj interface{}, mode base.SpatialMode) {
	this.obj = obj
	this.mode = mode
}

// 根据查询对象和查询模式，返回空间索引的编码；用于数据库存储库（空间索引与索引编码绑定）
func (this *SpatailQuery) QueryCodes(idx index.SpatialIndexDB) (codes []int32) {
	if this.obj == nil {
		return nil
	}
	// 查询阶段，关键还是用边框来解决问题
	bounds := getBounds(this.obj)
	if bounds != nil {
		disjoint := this.mode.IsDisjoint()
		if disjoint {
			codes = idx.QueryNoCoveredDB(bounds.GetBounds())
		} else {
			codes = idx.QueryDB(bounds.GetBounds())
		}
	}
	return
}

// 根据查询对象和查询模式，返回空间索引的编码；用于文件型存储库（空间索引与id绑定）
func (this *SpatailQuery) QueryIds(idx index.SpatialIndex) (ids []int64) {
	if this.obj == nil {
		return nil
	}
	// 查询阶段，关键还是用边框来解决问题
	bounds := getBounds(this.obj)
	if bounds != nil {
		disjoint := this.mode.IsDisjoint()
		if disjoint {
			ids = idx.QueryNoCovered(bounds.GetBounds())
		} else {
			ids = idx.Query(bounds.GetBounds())
		}
	}
	return
}

func getBounds(obj interface{}) base.Bounds {
	var bounds base.Bounds
	if bbox, ok := obj.(base.Rect2D); ok {
		bounds = &bbox
	} else if geo, ok := obj.(geometry.Geometry); ok {
		bounds = geo
	} else if feaset, ok := obj.(Featureset); ok {
		bounds = feaset
	}
	return bounds
}

// 判断空间关系是否匹配通过
func (this *SpatailQuery) Match(geo geometry.Geometry) bool {
	if bbox, ok := this.obj.(base.Rect2D); ok {
		return BboxMatchGeo(bbox, this.mode, geo)
		// var geoPolygon geometry.GeoPolygon
		// geoPolygon.Make(bbox)
		// var relation geometry.GeoRelation
		// relation.A = &geoPolygon
		// relation.B = geo
		// return relation.IsMatch(this.mode)
	} else if geoa, ok := this.obj.(geometry.Geometry); ok {
		var relation geometry.GeoRelation
		relation.A = geoa
		relation.B = geo
		return relation.IsMatch(this.mode)
	} // else todo, such as featureset ...
	return true
}

// 矩形查询几何对象
func BboxMatchGeo(bbox base.Rect2D, mode base.SpatialMode, geo geometry.Geometry) bool {
	switch mode {
	case base.BBoxIntersects:
		return bbox.IsIntersects(geo.GetBounds())
	// 这几个可调换a/b顺序: Intersects,Disjoint,Equal,Overlap,Touch,Cross(线线)
	case base.Intersects, base.Undefined, base.Disjoint, base.Equals, base.Overlaps, base.Touches:
		return geo.IsRelate(mode, bbox)
	// 剩下的不能换顺序
	case base.Crosses:
		// 矩形（面）似乎没有办法做crosses
		return false
	case base.Contains:
		// 包含的，用bounds就ok
		return bbox.IsContains(geo.GetBounds())
	case base.Covers:
		// 覆盖的，用bounds就ok
		return bbox.IsCovers(geo.GetBounds())
	case base.Within:
		// bbox在 geo的内部（且边界不接触）
		return geo.IsRelate(base.Contains, bbox)
	case base.CoveredBy:
		// bbox在 geo的内部（且边界可接触）
		return geo.IsRelate(base.Covers, bbox)
	default:
		var im base.D9IM
		im.Init(string(mode))
		// 把矩阵转置一下，然后颠倒 a/b的顺序
		im.Invert()
		return geo.IsRelate(base.SpatialMode(im.String()), bbox)
	}
	return true
}
