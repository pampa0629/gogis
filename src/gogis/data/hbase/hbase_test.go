package hbase

import (
	"gogis/base"

	"gogis/data"
	_ "gogis/data/shape"
	"testing"
)

func TestConvertShp2Hbase(t *testing.T) {
	tr := base.NewTimeRecorder()

	title := "chinapnt_84" // JBNTBHTB chinapnt_84

	fromParams := data.NewConnParams()
	fromParams["filename"] = "c:/temp/" + title + ".shp"
	fromParams["type"] = string(data.StoreShape)

	toParams := data.NewConnParams()
	toParams["address"] = "localhost:2181"
	toParams["type"] = string(data.StoreHbase)

	var cvt data.Converter
	cvt.Convert(fromParams, title, toParams)

	tr.Output("convert")
}
