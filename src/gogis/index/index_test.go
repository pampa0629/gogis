package index

import (
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"os"
	"testing"
)

func TestSlice(t *testing.T) {
	s := make([]int, 1)
	s[0] = 111
	s = append(s[:0], s[0+1:]...) // 删除中间1个元素
	fmt.Println("slice:", s)
}

func TestRTreeIndex(t *testing.T) {
	geos, bbox, count := makeGeos33()
	var rindex RTreeIndex
	rindex.Init(bbox, count)
	// rindex.objCount = 3
	RTREE_OBJ_COUNT = 3
	rindex.BuildByGeos(geos)
	rindex.WholeString()
	if !rindex.Check() {
		t.Errorf("build RTreeIndex check false")
	}

	filename := "c:/temp/test.six"
	f, _ := os.Create(filename)
	rindex.Save(f)
	f.Close()

	six, _ := os.Open(filename)
	var rindex2 RTreeIndex
	rindex2.Load(six)
	rindex2.WholeString()
	if !rindex2.Check() {
		t.Errorf("loaded RTreeIndex check false")
	}
	return
}

func makeGeos33() (geos []geometry.Geometry, bbox base.Rect2D, count int64) {
	length, width := 3, 3
	geos = make([]geometry.Geometry, length*width)
	bbox.Init()

	for i := 0; i < length; i++ {
		for j := 0; j < width; j++ {
			geoPoint := new(geometry.GeoPoint)
			geoPoint.Point2D = base.Point2D{float64(i), float64(j)}
			geos[i*width+j] = geoPoint
			bbox.Union(geoPoint.GetBounds())
		}
	}
	return geos, bbox, int64(length * width)
}

func makeGeos24() (geos []geometry.Geometry, bbox base.Rect2D, count int64) {
	count = 8
	geos = make([]geometry.Geometry, count)
	bbox.Init()

	geos[0] = getGeoPoint(0, 0)
	geos[1] = getGeoPoint(1, 0)
	geos[2] = getGeoPoint(0, 2)
	geos[3] = getGeoPoint(1, 2)
	geos[4] = getGeoPoint(0, 6)
	geos[5] = getGeoPoint(1, 6)
	geos[6] = getGeoPoint(0, 10)
	geos[7] = getGeoPoint(1, 10)

	for _, v := range geos {
		bbox.Union(v.GetBounds())
	}

	return geos, bbox, count
}

func getGeoPoint(x, y float64) *geometry.GeoPoint {
	geoPoint := new(geometry.GeoPoint)
	geoPoint.Point2D = base.Point2D{x, y}
	return geoPoint
}
