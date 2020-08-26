package geometry

import "gogis/base"

type GeoPoint struct {
	Point2D
}

func (this *GeoPoint) Type() GeoType {
	return TGeoPoint
}

func (this *GeoPoint) GetBounds() base.Rect2D {
	return base.NewRect2D(this.X, this.Y, this.X, this.Y)
}

// ID() int64
// Draw() // todo 绘制参数
// From(data []byte, mode GeoMode) bool
// To(mode GeoMode) []byte
