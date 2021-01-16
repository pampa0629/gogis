package geometry

import (
	"gogis/draw"

	"github.com/tidwall/mvt"
)

type MvtTools struct {
	// GeoType2Mvt(geotype GeoType) GeometryType
}

func GeoType2Mvt(geotype GeoType) mvt.GeometryType {
	switch geotype {
	case TGeoPoint:
		return mvt.Point
	case TGeoPolyline:
		return mvt.LineString
	case TGeoPolygon:
		return mvt.Polygon
	}
	return mvt.Unknown
}

func Geo2Mvt(geo Geometry, mvtFea *mvt.Feature, canvas *draw.Canvas) {
	switch geo.Type() {
	case TGeoPoint:
		geoPoint2Mvt(geo.(*GeoPoint), mvtFea, canvas)
	case TGeoPolyline:
		geoPolyline2Mvt(geo.(*GeoPolyline), mvtFea, canvas)
	case TGeoPolygon:
		geoPolygon2Mvt(geo.(*GeoPolygon), mvtFea, canvas)
	}
}

func geoPoint2Mvt(geo *GeoPoint, mvtFea *mvt.Feature, canvas *draw.Canvas) {
	pnt := canvas.Params.Forward(geo.Point2D)
	mvtFea.MoveTo(float64(pnt.X), float64(pnt.Y))
}

func geoPolyline2Mvt(geo *GeoPolyline, mvtFea *mvt.Feature, canvas *draw.Canvas) {
	for _, v := range geo.Points {
		pnts := canvas.Params.Forwards(v)
		geoline2Mvt(pnts, mvtFea)
	}
}

// todo 节点顺序问题；mvt如何支持多个岛洞的问题
func geoPolygon2Mvt(geo *GeoPolygon, mvtFea *mvt.Feature, canvas *draw.Canvas) {
	for _, v := range geo.Points {
		for _, vv := range v {
			pnts := canvas.Params.Forwards(vv)
			geoline2Mvt(pnts, mvtFea)
			mvtFea.ClosePath()
		}
	}
}

func geoline2Mvt(pnts []draw.Point, mvtFea *mvt.Feature) {
	mvtFea.MoveTo(float64(pnts[0].X), float64(pnts[0].Y))
	for j := 1; j < len(pnts); j++ {
		mvtFea.LineTo(float64(pnts[j].X), float64(pnts[j].Y))
	}
}
