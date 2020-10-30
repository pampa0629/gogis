package data

import "testing"

func TestOpenShape(t *testing.T) {
	var shpStore = new(ShapeStore)
	params := NewConnParams()
	params["filename"] = "./testdata/chinapnt.shp"
	ok, err := shpStore.Open(params)
	if !ok || err != nil {
		t.Errorf(err.Error())
	}
	shp, _ := shpStore.GetFeasetByNum(0)
	if shp.Count() != 2391 {
		t.Errorf("对象数量不对")
	}
}
