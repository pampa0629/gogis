package data

import (
	"gogis/base"
	"gogis/geometry"
)

type SpatialIndex interface {
	Init(bbox base.Rect2D, num int)
	BuildByGeos(geometrys []*geometry.Geometry)
	BuildByFeas(features []*Feature)
	Clear()
}
