package geometry

import "gogis/base"

type GeoPolyline struct {
	Points [][]base.Point2D
	BBox   base.Rect2D
}

func (this *GeoPolyline) Type() GeoType {
	// 	if len(this.Points) == 1 {
	// 		return TGeoLineString
	// 	} else if len(this.Points) > 1 {
	// 		return TGeoMultiLineString
	// 	}
	return TGeoPolygon
}

func (this *GeoPolyline) GetBounds() base.Rect2D {
	return this.BBox
}

func (this *GeoPolyline) ComputeBounds() base.Rect2D {
	this.BBox.Init()
	for _, points := range this.Points {
		this.BBox.Union(base.ComputeBounds(points))
	}
	return this.BBox
}

// ID() int64
// Draw() // todo 绘制参数
// From(data []byte, mode GeoMode) bool
// To(mode GeoMode) []byte
