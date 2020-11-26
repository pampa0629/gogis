package data

import (
	"gogis/base"
	"gogis/geometry"
)

type SpatialIndex interface {
	Init(bbox base.Rect2D, num int)
	BuildByGeos(geometrys []geometry.Geometry)
	BuildByFeas(features []Feature)

	// 范围查询，返回id数组
	Query(bbox base.Rect2D) []int

	Clear()
}
