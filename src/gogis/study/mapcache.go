package main

import (
	"fmt"
	"gogis/base"
	"gogis/data/shape"
	"gogis/draw"
	"gogis/mapping"
)

func main() {
	// testCache()
	testCacheRaster()
	fmt.Println("DONE!")
}

func testCacheRaster() {
	tr := base.NewTimeRecorder()
	path := "C:/BigData/10_Data/testimage/image2/"
	// path := "C:/BigData/10_Data/images/imagebig2/"
	gmap := mapping.NewMap()
	gmap.Open(path + "image2.gmp")
	tr.Output("open map")
	mapTile := mapping.NewMapTile(gmap, mapping.Epsg4326)

	mapTile.Cache("c:/temp/cache/", gmap.Name, draw.TypePng, PbCache)
	tr.Output("map cache:" + gmap.Name)
}

func testCache() {
	tr := base.NewTimeRecorder()

	gmap := startMap()
	mapTile := mapping.NewMapTile(gmap, mapping.Epsg4326)
	mapTile.Cache("c:/temp/cache/", gmap.Name, draw.TypePng, PbCache)

	tr.Output("cache map:" + gmap.Name)
}

var cancel = false

func PbCache(title, sub string, no, count int, step, total int64, cost, estimate int) bool {
	if (total > 100 && step%(total/100) == 0) || total <= 100 {
		fmt.Println(title, sub, no, count, step, total, cost, estimate)
		// if !cancel {
		// 	fmt.Println("PbCache cancel")
		// 	cancel = true
		// 	return true
		// }
	}
	return false
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
	gmap.AddFeatureLayer(feaset, nil)
	return gmap
}
