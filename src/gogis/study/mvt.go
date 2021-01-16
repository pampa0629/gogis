package main

import (
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/mapping"
	"os"

	// tidwall/mvt
	"github.com/tidwall/mvt"
)

func mvtmain() {
	testMapTile()
	// testTile()
	fmt.Println("gg DONE")
}

func testMapTile() {
	var gTitle = "chinapnt_84" // insurance chinapnt_84
	// var gTitle = "DLTB"
	// var gTitle = "JBNTBHTB"

	tr := base.NewTimeRecorder()
	var store data.ShpmemStore
	params := data.NewConnParams()
	params["filename"] = "C:/temp/" + gTitle + ".shp"
	store.Open(params)
	// feaset, _ := sqlDB.GetFeasetByNum(0)
	feaset, _ := store.GetFeasetByName(gTitle)
	feaset.Open()
	tr.Output("open shp by memery")

	gmap := mapping.NewMap()
	// var theme mapping.RangeTheme // UniqueTheme
	// gmap.AddLayer(feaset, &theme)
	gmap.AddLayer(feaset, nil)
	gmap.Prepare(256, 256)
	// gmap.Zoom(5)
	gmap.Draw()
	// 输出图片文件
	gmap.Output2File("C:/temp/"+gTitle+".jpg", "jpg")
	data, _ := gmap.OutputMvt()
	// fmt.Println("mvt data:", data)
	f, _ := os.Create("C:/temp/" + gTitle + ".mvt")
	f.Write(data)
	f.Close()
	// mapfile := gPath + gTitle + "." + base.EXT_MAP_FILE
	// gmap.Save(mapfile)

	tr.Output("draw map")
}

func testTile() {
	var tile mvt.Tile
	l := tile.AddLayer("triforce")
	f := l.AddFeature(mvt.Polygon)

	f.MoveTo(128, 96)
	f.LineTo(148, 128)
	f.LineTo(108, 128)
	f.LineTo(128, 96)
	f.ClosePath()

	f.MoveTo(148, 128)
	f.LineTo(168, 160)
	f.LineTo(128, 160)
	f.LineTo(148, 128)
	f.ClosePath()

	f.MoveTo(108, 128)
	f.LineTo(128, 160)
	f.LineTo(88, 160)
	f.LineTo(108, 128)
	f.ClosePath()

	data := tile.Render()
	fmt.Println("string:", string(data))
	fmt.Println("data:", data)

}
