package geometry

import (
	"fmt"
	"gogis/base"
	"testing"
)

func TestRelPolygonPolygon(t *testing.T) {
	var a GeoPolygon
	var ra base.Rect2D
	var b GeoPolygon
	var rb base.Rect2D
	var relation GeoRelation
	var im base.D9IM
	relation.A = &a
	relation.B = &b
	rb.Max.X = 100
	rb.Max.Y = 100
	b.Make(rb)

	ra.Min.X = -100
	ra.Min.Y = -100
	ra.Max.X = -10
	ra.Max.Y = -10
	a.Make(ra)
	im = relation.CalcRelateIM()
	if !relation.IsDisjoint() {
		fmt.Println("im1:", im.String())
		t.Errorf("面&面错误1")
	}

	ra.Max.X = 0
	ra.Max.Y = 0
	a.Make(ra)
	im = relation.CalcRelateIM()
	if !im.Match("FF2F01212") || !relation.IsTouches() {
		fmt.Println("im2:", im.String())
		t.Errorf("面&面错误2")
	}

	ra.Max.X = 50
	a.Make(ra)
	im = relation.CalcRelateIM()
	if !im.Match("FF2F11212") || !relation.IsTouches() {
		fmt.Println("im3:", im.String())
		t.Errorf("面&面错误3")
	}

	ra.Max.X = 50
	ra.Max.Y = 50
	a.Make(ra)
	im = relation.CalcRelateIM()
	if !relation.IsOverlaps() {
		fmt.Println("im4:", im.String())
		t.Errorf("面&面错误4")
	}

	ra.Min.X = 10
	ra.Min.Y = 10
	ra.Max.X = 90
	ra.Max.Y = 90
	a.Make(ra)
	im = relation.CalcRelateIM()
	if !relation.IsWithin() {
		fmt.Println("im5:", im.String())
		t.Errorf("面&面错误5")
	}

	ra.Min.X = -10
	ra.Min.Y = -10
	ra.Max.X = 110
	ra.Max.Y = 110
	a.Make(ra)
	im = relation.CalcRelateIM()
	if !relation.IsContains() {
		fmt.Println("im6:", im.String())
		t.Errorf("面&面错误6")
	}

	ra.Min.X = 0
	ra.Min.Y = 0
	ra.Max.X = 90
	ra.Max.Y = 90
	a.Make(ra)
	im = relation.CalcRelateIM()
	if !relation.IsCoveredBy() {
		fmt.Println("im7:", im.String())
		t.Errorf("面&面错误7")
	}

	ra.Max.X = 110
	ra.Max.Y = 110
	a.Make(ra)
	im = relation.CalcRelateIM()
	if !relation.IsCovers() {
		fmt.Println("im8:", im.String())
		t.Errorf("面&面错误8")
	}

	ra.Max.X = 100
	ra.Max.Y = 100
	a.Make(ra)
	im = relation.CalcRelateIM()
	if !relation.IsEquals() {
		fmt.Println("im9:", im.String())
		t.Errorf("面&面错误9")
	}
}

func TestRelPolylinePolygon(t *testing.T) {
	var a GeoPolyline
	a.Points = make([][]base.Point2D, 1)
	a.Points[0] = make([]base.Point2D, 2)
	a.Points[0][1].X = 100
	a.ComputeBounds()

	var b GeoPolygon
	var rect base.Rect2D
	rect.Max.X = 100
	rect.Max.Y = 100
	b.Make(rect)

	var relation GeoRelation
	var im base.D9IM
	relation.A = &a
	relation.B = &b
	im = relation.CalcRelateIM()
	if !im.Match("F1FF0F212") || !relation.IsTouches() {
		fmt.Println("im1:", im.String())
		t.Errorf("线&面错误1")
	}

	a.Points[0][1].X = -100
	a.ComputeBounds()
	im = relation.CalcRelateIM()
	if !im.Match("FF1F00212") || !relation.IsTouches() {
		fmt.Println("im2:", im.String())
		t.Errorf("线&面错误2")
	}

	a.Points[0][0].X = -10
	a.ComputeBounds()
	im = relation.CalcRelateIM()
	if !relation.IsDisjoint() {
		fmt.Println("im3:", im.String())
		t.Errorf("线&面错误3")
	}

	a.Points[0][0].X = -10
	a.Points[0][0].Y = 10
	a.Points[0][1].X = 50
	a.Points[0][1].Y = 50
	a.ComputeBounds()
	im = relation.CalcRelateIM()
	if !relation.IsCrosses() {
		fmt.Println("im4:", im.String())
		t.Errorf("线&面错误4")
	}

	a.Points[0][0].X = 10
	a.ComputeBounds()
	im = relation.CalcRelateIM()
	if !relation.IsWithin() {
		fmt.Println("im5:", im.String())
		t.Errorf("线&面错误5")
	}
}

func TestRelPntPolyline(t *testing.T) {
	var a GeoPoint
	var b GeoPolyline
	b.Points = make([][]base.Point2D, 1)
	b.Points[0] = make([]base.Point2D, 2)
	b.Points[0][1].X = 100
	b.ComputeBounds()
	var relation GeoRelation
	relation.A = &a
	relation.B = &b
	im := relation.CalcRelateIM()
	if !im.Match("F0F******") || !relation.IsTouches() {
		fmt.Println("im:", im.String())
		t.Errorf("点&线错误1")
	}
	a.Point2D.X = 50
	im = relation.CalcRelateIM()
	if !im.Match("0FF******") || !relation.IsWithin() {
		fmt.Println("im:", im.String())
		t.Errorf("点&线错误2")
	}

	a.Point2D.Y = 50
	im = relation.CalcRelateIM()
	if !im.Match("FF0******") || !relation.IsDisjoint() {
		fmt.Println("im:", im.String())
		t.Errorf("点&线错误3")
	}
}

func TestRelPntPnt(t *testing.T) {
	var a, b GeoPoint
	var relation GeoRelation
	relation.A = &a
	relation.B = &b
	im := relation.CalcRelateIM()
	if !im.Match("0*F***F*2") || !relation.IsEquals() {
		fmt.Println("im:", im.String())
		t.Errorf("点&点错误1")
	}
	a.Point2D.X = 100
	im = relation.CalcRelateIM()
	if !im.Match("F*0***0*2") || !relation.IsDisjoint() {
		fmt.Println("im:", im.String())
		t.Errorf("点&点错误1")
	}
}

func TestRelPntPolygon(t *testing.T) {
	// 点和面
	{
		var geoPoint GeoPoint
		// geoPoint.Point2D.X = 0
		var geoPolygon GeoPolygon
		var rect base.Rect2D
		rect.Max.X = 100
		rect.Max.Y = 100
		geoPolygon.Make(rect)
		var relation GeoRelation
		relation.A = &geoPoint
		relation.B = &geoPolygon
		im := relation.CalcRelateIM()
		if !im.Match("F0F***212") {
			fmt.Println("im:", im.String())
			t.Errorf("点&面错误1")
		}

		geoPoint.X = -1
		im = relation.CalcRelateIM()
		if !im.Match("FF0***212") {
			fmt.Println("im:", im.String())
			t.Errorf("点&面错误2")
		}

		geoPoint.X = 50
		geoPoint.Y = 50
		im = relation.CalcRelateIM()
		if !im.Match("0FF***212") {
			fmt.Println("im:", im.String())
			t.Errorf("点&面错误3")
		}
	}

}
