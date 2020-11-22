package geometry

import (
	"fmt"
	"gogis/base"
	"testing"
)

func TestWkbPoint(t *testing.T) {
	var pnt, pnt2 GeoPoint
	pnt.X = 121.1
	pnt.Y = 42.4
	data := pnt.To(WKB)
	pnt2.From(data, WKB)
	if pnt.X != pnt.X || pnt.Y != pnt.Y {
		t.Errorf("几何点对象WKB错误")
	}
}

func TestWkbPolyine(t *testing.T) {
	var geo, geo2 GeoPolyline
	geo.Points = make([][]base.Point2D, 1)
	geo.Points[0] = make([]base.Point2D, 2)
	geo.Points[0][0].X = 122.1
	geo.Points[0][0].Y = 132.1
	geo.Points[0][1].X = 142.1
	geo.Points[0][1].Y = 152.1

	data := geo.To(WKB)
	geo2.From(data, WKB)
	for i, v := range geo.Points {
		for ii, vv := range v {
			if vv.X != geo2.Points[i][ii].X || vv.Y != geo2.Points[i][ii].Y {
				t.Errorf("几何线对象WKB错误")
			}
		}
	}
}

func TestWkbPolygon(t *testing.T) {
	var geo, geo2 GeoPolygon
	geo.Points = make([][][]base.Point2D, 2)
	geo.Points[0] = make([][]base.Point2D, 1)
	geo.Points[0][0] = make([]base.Point2D, 3)
	geo.Points[0][0][0].X = 102.1
	geo.Points[0][0][0].Y = 132.1
	geo.Points[0][0][1].X = 142.1
	geo.Points[0][0][1].Y = 152.1
	geo.Points[0][0][2].X = 102.1
	geo.Points[0][0][2].Y = 132.1

	geo.Points[1] = make([][]base.Point2D, 1)
	geo.Points[1][0] = make([]base.Point2D, 3)
	geo.Points[1][0][0].X = 102.1
	geo.Points[1][0][0].Y = 132.1
	geo.Points[1][0][1].X = 142.1
	geo.Points[1][0][1].Y = 152.1
	geo.Points[1][0][2].X = 102.1
	geo.Points[1][0][2].Y = 132.1

	fmt.Println("geo's len:", len(geo.Points))

	data := geo.To(WKB)
	geo2.From(data, WKB)
	fmt.Println(len(data), data)
	fmt.Println(geo2)
	for i, v := range geo.Points {
		for ii, vv := range v {
			for iii, vvv := range vv {
				if vvv.X != geo2.Points[i][ii][iii].X || vvv.Y != geo2.Points[i][ii][iii].Y {
					t.Errorf("几何线对象WKB错误")
					// t.Error(geo.Points)
					// t.Error(geo2.Points)
				}
			}
		}
	}
}
