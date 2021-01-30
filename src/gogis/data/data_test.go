package data

import (
	"fmt"
	"gogis/algorithm"
	"gogis/base"
	"strconv"
	"testing"
)

func TestQuery(t *testing.T) {
	// var store = new(ShpmemStore)
	var store = new(SqliteStore)
	params := NewConnParams()
	// params["filename"] = "./testdata/chinapnt.shp"
	params["filename"] = "./testdata/chinapnt_84.sqlite"
	ok, err := store.Open(params)
	if !ok || err != nil {
		t.Errorf(err.Error())
	}
	temp, _ := store.GetFeasetByNum(0)
	feaset := temp.(*SqliteFeaset)
	// feaset := temp.(*ShpmemFeaset)
	feaset.Open()
	var def QueryDef
	def.Fields = []string{"POPU", "POP_COU"}
	def.Where = "POPU>100 or POPU<80 and POP_COU>10"
	// def.Where = "(Popu>10 or Pop_cou>10) or((a<=11) and (b>0) or c!=1)"
	def.SpatialMode = algorithm.Disjoint // Intersects Within Disjoint "[T***F*FF*]"
	def.SpatialObj = feaset.bbox.Scale(0.5)
	feait := feaset.QueryByDef(def)
	feait.PrepareBatch(3000)
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
	fromParams["type"] = string(StoreShapeMemory)

	toParams := NewConnParams()
	toParams["addresses"] = "http://localhost:9200"
	toParams["type"] = string(StoreES)

	var cvt Converter
	cvt.Convert(fromParams, title, toParams)

	tr.Output("convert")
}

func TestConvertShp2Hbase(t *testing.T) {
	tr := base.NewTimeRecorder()

	title := "JBNTBHTB" // JBNTBHTB chinapnt_84

	fromParams := NewConnParams()
	fromParams["filename"] = "c:/temp/" + title + ".shp"
	fromParams["type"] = string(StoreShapeMemory)

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
	var store HbaseStore
	store.Open(params)
	store.DeleteFeaset(title)

	tr.Output("DeleteFeaset")
}

func TestOpenEs(t *testing.T) {
	var store = new(EsStore)
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
	var shpStore = new(ShapeStore)
	params := NewConnParams()
	params["filename"] = "./testdata/chinapnt.shp"
	ok, err := shpStore.Open(params)
	if !ok || err != nil {
		t.Errorf(err.Error())
	}
	shp, _ := shpStore.GetFeasetByNum(0)
	if shp.GetCount() != 2391 {
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
	feaset.Open()
	count := feaset.GetCount()
	if count != 2391 {
		t.Errorf("对象数量不对:" + strconv.Itoa(int(count)))
	}
}
