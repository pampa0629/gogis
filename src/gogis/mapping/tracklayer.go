package mapping

import (
	"gogis/base"
	"gogis/draw"
	"gogis/geometry"
)

// 图层类
type TrackLayer struct {
	style draw.Style
	geos  []geometry.Geometry
}

// todo 动态投影
func (this *TrackLayer) Draw(canvas draw.Canvas, proj *base.ProjInfo) {
	canvas.SetStyle(this.style)
	for _, v := range this.geos {
		drawGeo, ok := v.(draw.DrawCanvas)
		if ok {
			drawGeo.Draw(canvas)
		}
	}
}
