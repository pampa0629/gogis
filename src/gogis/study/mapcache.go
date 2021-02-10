package main

import (
	"gogis/base"
	"gogis/data/shape"
	"gogis/draw"
	"gogis/mapping"
)

func mcmain() {
	testCache()
}

func testCache() {
	tr := base.NewTimeRecorder()

	gmap := startMap()
	mapTile := mapping.NewMapTile(gmap, mapping.Epsg4326)
	mapTile.Cache("c:/temp/cache/", gmap.Name, draw.TypePng)

	tr.Output("cache map:" + gmap.Name)
}

func startMap() *mapping.Map {

	var gPath = "C:/temp/"

	// var gTitle = "chinapnt_84"

	// var gTitle = "DLTB"

	var gTitle = "JBNTBHTB"

	var gExt = ".shp"
	var filename = gPath + gTitle + gExt

	// 打开shape文件
	feaset := shape.OpenShape(filename, true, []string{})
	// // 创建地图
	gmap := mapping.NewMap()
	gmap.AddLayer(feaset, nil)
	return gmap
}
