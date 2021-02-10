package sqlite

import (
	"gogis/base"
	"gogis/data"
	_ "gogis/data/shape"
	"strconv"
	"testing"
)

func TestOpenSpatialite(t *testing.T) {
	store := data.NewDatastore(data.StoreSqlite)
	params := data.NewConnParams()
	params["filename"] = "c:/temp/chinapnt_84.sqlite"
	ok, err := store.Open(params)
	if !ok || err != nil {
		t.Errorf(err.Error())
	}
	feaset := store.GetFeasetByNum(0)
	feaset.Open()
	count := feaset.GetCount()
	if count != 2391 {
		t.Errorf("对象数量不对:" + strconv.Itoa(int(count)))
	}
}

func TestShp2Sqlite(t *testing.T) {
	tr := base.NewTimeRecorder()
	title := "point2" // JBNTBHTB chinapnt_84 point2 DLTB

	fromParams := data.NewConnParams()
	fromParams["filename"] = "c:/temp/" + title + ".shp"
	fromParams["type"] = string(data.StoreShape)

	toParams := data.NewConnParams()
	toParams["filename"] = "c:/temp/" + title + ".sqlite"
	toParams["type"] = string(data.StoreSqlite)

	var cvt data.Converter
	cvt.Convert(fromParams, title, toParams)
	tr.Output("convert")
}
