package main

import (
	"gogis/base"
	"gogis/data"
	"gogis/mapping"
)

func main() {
	testCache()
}

func testCache() {
	tr := base.NewTimeRecorder()

	gmap := startMap()
	mapTile := mapping.NewMapTile(gmap, mapping.Epsg4326)
	mapTile.Cache("c:/temp/cache/", gmap.Name)

	tr.Output("cache map:" + gmap.Name)
}

func startMap() *mapping.Map {

	var gPath = "C:/temp/"

	// var gTitle = "chinapnt_84"

	var gTitle = "DLTB"

	// var gTitle = "JBNTBHTB"

	var gExt = ".shp"
	var filename = gPath + gTitle + gExt

	// 打开shape文件
	feaset := data.OpenShape(filename)
	// // 创建地图
	gmap := mapping.NewMap()
	gmap.AddLayer(feaset)
	return gmap
}
