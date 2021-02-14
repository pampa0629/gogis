package data

import (
	"fmt"
	"gogis/base"

	// _ "gogis/data/sqlite"
	"testing"
)

func TestConvertLine2Pnt(t *testing.T) {
	tr := base.NewTimeRecorder()

	title := "line" // JBNTBHTB chinapnt_84

	fromParams := NewConnParams()
	fromParams["filename"] = "c:/temp/" + title + ".udbx" // line.udbx
	fromParams["type"] = string(StoreSqlite)
	fromStore := NewDatastore(StoreSqlite)
	fromFeaset := fromStore.GetFeasetByNum(0)
	fromFeaset.Open()

	toParams := NewConnParams()
	toParams["filename"] = "c:/temp/" + "point" + ".sqlite"
	toParams["type"] = string(StoreSqlite)
	toStore := NewDatastore(StoreSqlite)

	var cvt Converter
	cvt.Polyline2Point(fromFeaset, toStore, "point")
	tr.Output("Polyline2Point")
}

func TestQuery(t *testing.T) {
	// var store = new(ShpmemStore)
	// var store = new(SqliteStore)
	store := NewDatastore(StoreSqlite)
	params := NewConnParams()
	// params["filename"] = "./testdata/chinapnt.shp"
	params["filename"] = "./testdata/chinapnt_84.sqlite"
	ok, err := store.Open(params)
	if !ok || err != nil {
		t.Errorf(err.Error())
	}
	feaset := store.GetFeasetByNum(0)
	// feaset := temp.(*SqliteFeaset)
	// feaset := temp.(*ShpmemFeaset)
	feaset.Open()
	var def QueryDef
	def.Fields = []string{"POPU", "POP_COU"}
	def.Where = "POPU>100 or POPU<80 and POP_COU>10"
	// def.Where = "(Popu>10 or Pop_cou>10) or((a<=11) and (b>0) or c!=1)"
	def.SpatialMode = base.Disjoint // Intersects Within Disjoint "[T***F*FF*]"
	def.SpatialObj = feaset.GetBounds().Scale(0.5)
	feait := feaset.Query(&def)
	feait.BeforeNext(3000)
	feas, ok := feait.BatchNext(0)
	fmt.Println("count:", len(feas))
	// for _, v := range feas {
	// 	fmt.Println(v.Geo.GetID(), v.Atts)
	// }
}

func TestConvertShp2Es(t *testing.T) {
	tr := base.NewTimeRecorder()

	title := "insurance" // JBNTBHTB chinapnt_84

	fromParams := NewConnParams()
	fromParams["filename"] = "c:/temp/" + title + ".shp"
	fromParams["type"] = string(StoreShape)

	toParams := NewConnParams()
	toParams["addresses"] = "http://localhost:9200"
	toParams["type"] = string(StoreES)

	var cvt Converter
	cvt.Convert(fromParams, title, toParams)

	tr.Output("convert")
}

func TestConvertShp2Hbase(t *testing.T) {
	tr := base.NewTimeRecorder()

	title := "chinapnt_84" // JBNTBHTB chinapnt_84

	fromParams := NewConnParams()
	fromParams["filename"] = "c:/temp/" + title + ".shp"
	fromParams["type"] = string(StoreShape)

	toParams := NewConnParams()
	toParams["address"] = "localhost:2181"
	toParams["type"] = string(StoreHbase)

	var cvt Converter
	cvt.Convert(fromParams, title, toParams)

	tr.Output("convert")
}

func TestDelete(t *testing.T) {
	tr := base.NewTimeRecorder()

	title := "JBNTBHTB" // JBNTBHTB

	params := NewConnParams()
	params["address"] = "localhost:2181"
	params["type"] = string(StoreHbase)
	// var store HbaseStore
	store := NewDatastore(StoreHbase)
	store.Open(params)
	store.DeleteFeaset(title)

	tr.Output("DeleteFeaset")
}

func TestOpenEs(t *testing.T) {
	// var store = new(EsStore)
	store := NewDatastore(StoreES)
	params := NewConnParams()
	params["addresses"] = []string{"http://localhost:9200"}
	ok, err := store.Open(params)
	if !ok || err != nil {
		t.Errorf(err.Error())
	}
	feaset, _ := store.GetFeasetByName("geodata")
	fmt.Println(feaset.GetName())

}

func TestOpenShape(t *testing.T) {
	// var shpStore = new(ShapeStore)
	store := NewDatastore(StoreShape)
	params := NewConnParams()
	params["filename"] = "./testdata/chinapnt.shp"
	ok, err := store.Open(params)
	if !ok || err != nil {
		t.Errorf(err.Error())
	}
	shp := store.GetFeasetByNum(0)
	if shp.GetCount() != 2391 {
		t.Errorf("对象数量不对")
	}
}

func TestMosaic(t *testing.T) {
	filename := "C:/temp/filelist.txt"
	gmrname := "C:/temp/filelist.gmr"
	var raset, raset2 MosaicRaset
	raset.Open(filename)
	raset.Save(filename)
	raset2.Open(gmrname)
}
