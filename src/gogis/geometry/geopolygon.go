package geometry

import "gogis/base"

type GeoPolygon struct {
	Points [][][]base.Point2D // []: 点串 [][]: 带洞的单个多边形 [][][] 带岛的多边形
	BBox   base.Rect2D
}

func (this *GeoPolygon) Type() GeoType {
	if len(this.Points) == 1 {
		return TGeoPolygon
	} else if len(this.Points) > 1 {
		return TGeoMultiPolygon
	}
	return TGeoEmpty
}

func (this *GeoPolygon) GetBounds() base.Rect2D {
	return this.BBox
}

func (this *GeoPolygon) ComputeBounds() base.Rect2D {
	this.BBox.Init()
	for _, polygon := range this.Points {
		// 只有第一个是岛，需要计算bounds；后面有也是洞，可以不理会
		if len(polygon) >= 1 {
			this.BBox.Union(base.ComputeBounds(polygon[0]))
		}
	}
	return this.BBox
}

// ID() int64
// Draw() // todo 绘制参数
// From(data []byte, mode GeoMode) bool
// To(mode GeoMode) []byte
