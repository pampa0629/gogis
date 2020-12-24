package data

import (
	"strconv"
	"testing"
)

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

func TestOpenSpatialite(t *testing.T) {
	var store = new(SqliteStore)
	params := NewConnParams()
	params["filename"] = "c:/temp/DLTB.sqlite"
	ok, err := store.Open(params)
	if !ok || err != nil {
		t.Errorf(err.Error())
	}
	feaset, _ := store.GetFeasetByNum(0)
	feaset.Open(feaset.GetName())
	count := feaset.Count()
	if count != 2391 {
		t.Errorf("对象数量不对:" + strconv.Itoa(int(count)))
	}
}
