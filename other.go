package main

import (
	"fmt"
	"os"

	"gogis/base"
	"gogis/data"
	"gogis/data/es"
	"gogis/data/hbase"
	_ "gogis/data/memory"
	"gogis/data/shape"
	"gogis/data/sqlite"
	"gogis/draw"
	"gogis/mapping"
	"gogis/server"
)

func testRest() {
	rootpath := "C:/zengzm/GitHub/gogis/" // os.Args[0]
	datapath := "./data/"
	cachepath := "./cache/"
	// if len(os.Args) >= 3 {
	// 	datapath = os.Args[1]
	// 	cachepath = os.Args[2]
	// }
	datapath = base.GetAbsolutePath(rootpath, datapath)
	cachepath = base.GetAbsolutePath(rootpath, cachepath)
	fmt.Println("app path:", rootpath, ", data path:", datapath, ", cache path:", cachepath)
	server.StartServer(datapath+"/gogis.gms", cachepath)
}

func testRasterRest() {
	// rootpath := "C:/zengzm/GitHub/gogis/" // os.Args[0]
	// datapath := "./data/"
	// cachepath := "./cache/"
	// // if len(os.Args) >= 3 {
	// // 	datapath = os.Args[1]
	// // 	cachepath = os.Args[2]
	// // }
	// datapath = base.GetAbsolutePath(rootpath, datapath)
	// cachepath = base.GetAbsolutePath(rootpath, cachepath)
	// fmt.Println("app path:", rootpath, ", data path:", datapath, ", cache path:", cachepath)
	// path := "C:/BigData/10_Data/testimage/image2/"
	path := "C:/BigData/10_Data/images/imagebig2/"
	server.StartServer(path+"gogis.gms", path+"cache")
}

var gPath = "C:/temp/"

// var gTitle = "railway" // insurance chinapnt_84 railway-x

// var gTitle = "DLTB" // Export_Output DLTB New_Region point2 railway

var gTitle = "JBNTBHTB-x"

var gExt = ".shp"

var filename = gPath + gTitle + gExt

func other() {
	// testRasterRest()
	// testRest()

	// testDrawMap()
	// testMapTile()

	// testShpQuery()

	// testShpMap()
	// testEsMap()
	// testSqliteMap()
	// testHbaseMap()

	// testLine2Point()

	// testDrawTiff()
	// buildMosaic()

	fmt.Println("DONE!")
	return
}

func testEsMap() {
	tr := base.NewTimeRecorder()
	var store es.EsStore
	params := data.NewConnParams()
	params["addresses"] = "http://localhost:9200"
	store.Open(params)
	feaset, _ := store.GetFeasetByName(gTitle)
	feaset.Open()
	tr.Output("open es db")

	gmap := mapping.NewMap(draw.Default)
	var theme mapping.GridTheme
	gmap.AddFeatureLayer(feaset, &theme)
	// gmap.AddGridTheme(feaset)
	gmap.Prepare(1024, 768)
	// gmap.Zoom(0.2)
	// gmap.PanMap(gmap.BBox.Dx()/20, gmap.BBox.Dy()/20)
	gmap.Draw()
	// 输出图片文件
	gmap.Output2File("C:/temp/"+gTitle+".jpg", "jpg")
	mapfile := "C:/temp/" + gTitle + ".gmp"
	gmap.Save(mapfile)
	var nmap mapping.Map
	nmap.Open(mapfile)

	tr.Output("draw es map")
}

func testHbaseMap() {
	tr := base.NewTimeRecorder()
	var store hbase.HbaseStore
	params := data.NewConnParams()
	params["address"] = "localhost:2181"
	store.Open(params)
	feaset, _ := store.GetFeasetByName(gTitle)
	feaset.Open()
	tr.Output("open hbase db")

	gmap := mapping.NewMap(draw.Default)
	gmap.AddFeatureLayer(feaset, nil)
	gmap.Prepare(1024, 768)
	// gmap.Zoom(2)
	// gmap.PanMap(gmap.BBox.Dx()/20, gmap.BBox.Dy()/20)
	gmap.Draw()
	// 输出图片文件
	gmap.Output2File("C:/temp/"+gTitle+".jpg", "jpg")

	tr.Output("draw hbase map")
}

func testSqliteMap() {
	tr := base.NewTimeRecorder()
	var sqlDB sqlite.SqliteStore
	params := data.NewConnParams()
	params["filename"] = "C:/temp/" + gTitle + ".udbx" // sqlite udbx
	sqlDB.Open(params)
	feaset := sqlDB.GetFeasetByNum(0)
	// feaset, _ := sqlDB.GetFeasetByName(gTitle)
	feaset.Open()
	// feaset = data.Cache(feaset, []string{})
	tr.Output("open sqlite db")

	gmap := mapping.NewMap(draw.Default)
	// var theme mapping.RangeTheme // UniqueTheme
	// gmap.AddLayer(feaset, &theme)
	gmap.AddFeatureLayer(feaset, nil)
	// gmap.Add
	gmap.Prepare(1024, 768)
	// gmap.Zoom(10)
	gmap.Draw()
	tr.Output("draw sqlite map")
	// 输出图片文件
	gmap.Output2File("C:/temp/"+gTitle+".jpg", "jpg")
	tr.Output("output")
	mapfile := gPath + gTitle + "." + base.EXT_MAP_FILE
	gmap.Save(mapfile)
	gmap.Save(mapfile) // 支持反复存储
	tr.Output("save map file")

}

func testShpQuery() {
	// tr := base.NewTimeRecorder()
	// var store shape.ShapeStore
	// params := data.NewConnParams()
	// params["filename"] = "C:/temp/" + gTitle + ".shp"
	// store.Open(params)
	// temp, _ := store.GetFeasetByNum(0)
	// feaset := temp.(*shape.ShapeFeaset)
	// feaset.Open()
	// fmt.Println("all count:", feaset.GetCount())
	// tr.Output("open")

	// var def data.QueryDef
	// // def.Fields = []string{"POPU", "POP_COU"}
	// // def.Where = "POPU>100 or POPU<80 and POP_COU>10"
	// // def.Where = "(Popu>10 or Pop_cou>10) or((a<=11) and (b>0) or c!=1)"
	// def.SpatialMode = base.Intersects // Intersects Within Disjoint "[T***F*FF*]"
	// def.SpatialObj = feaset.GetBounds().Scale(0.1)
	// feait := feaset.QueryByDef(def)
	// fmt.Println("query count:", feait.Count())
	// tr.Output("QueryByDef")

	// feait.PrepareBatch(int(feait.Count()))
	// feas, _ := feait.BatchNext(0)
	// fmt.Println("get count:", len(feas))
	// tr.Output("BatchNext")

	// bbox := feaset.GetBounds()
	// bbox.Extend((bbox.Dx() + bbox.Dy()) / -10.0)
	// fmt.Println("bbox:", bbox)
	// feait := feaset.QueryByBounds(bbox)

	// gmap := mapping.NewMap()
	// gmap.AddLayer(feaset, nil)
	// gmap.Prepare(1600, 1200)
	// tr.Output("new map")
	// gmap.Select(bbox)

	// tr.Output("select")
	// // gmap.Zoom(5)
	// gmap.Draw()
	// // 输出图片文件
	// gmap.Output2File("C:/temp/"+gTitle+".jpg", "jpg")
	// mapfile := gPath + gTitle + "." + base.EXT_MAP_FILE
	// gmap.Save(mapfile)
	// // nmap := mapping.NewMap()
	// // nmap.Open(mapfile)

	// tr.Output("draw map")
}

func testShpMap() {
	tr := base.NewTimeRecorder()
	var store shape.ShapeStore
	params := data.NewConnParams()
	params["filename"] = "C:/temp/" + gTitle + ".shp"
	// params["cache"] = true
	params["fields"] = []string{}
	store.Open(params)
	feaset := store.GetFeasetByNum(0)
	// feaset, _ := store.GetFeasetByName(gTitle)
	feaset.Open()
	// feaset = data.Cache(feaset, []string{})
	tr.Output("open shp ")

	gmap := mapping.NewMap(draw.Default)

	// var theme mapping.RangeTheme // UniqueTheme
	// gmap.AddLayer(feaset, &theme)

	path := "C:/BigData/10_Data/images/imagebig2/"
	// path := "C:/BigData/10_Data/testimage/image2/"
	var raset data.MosaicRaset
	raset.Open(path + "image2.gmr")

	gmap.AddRasterLayer(&raset)
	gmap.AddFeatureLayer(feaset, nil)
	gmap.Prepare(1600, 1200)

	// gmap.Proj = base.PrjFromEpsg(3857)
	// gmap.SetDynamicProj(true) // 动态投影
	// gmap.Zoom(5)

	gmap.Draw()

	// 输出图片文件
	gmap.Output2File("C:/temp/"+gTitle+".jpg", "jpg")
	tr.Output("draw map")
	mapfile := gPath + gTitle + "." + base.EXT_MAP_FILE
	gmap.Save(mapfile)

	nmap := mapping.NewMap(draw.Default)
	nmap.Open(mapfile)
	nmap.Prepare(1200, 900)
	nmap.Draw()
	nmap.Output2File("C:/temp/"+gTitle+"-2.jpg", "jpg")

	tr.Output("draw map")
}

func testMapTile() {
	tr := base.NewTimeRecorder()

	var gm base.GoMax
	gm.Init(500)
	for i := 0; i < 1000; i++ {
		gm.Add()
		go testMapTile2(nil, &gm)
	}
	gm.Wait()

	tr.Output("map tile")
}

func testMapTile2(gmap *mapping.Map, gm *base.GoMax) {

	gmap = mapping.NewMap(draw.Default)
	path := "C:/BigData/10_Data/testimage/image2/"
	// path := "C:/BigData/10_Data/images/imagebig2/"
	gmap.Open(path + "image2.gmp") //sqlite hbase "c:/temp/JBNTBHTB-hbase.gmp"

	maptile := mapping.NewMapTile(gmap, mapping.Epsg4326)
	tilename := gPath + "image2.png"
	// fmt.Println(tilename)
	data, _ := maptile.CacheOneTile2Bytes(4, 24, 5, draw.TypePng)
	w, _ := os.Create(tilename)
	w.Write(data)
	w.Close()

	if gm != nil {
		defer gm.Done()
	}
	return
}

func startMap() *mapping.Map {
	// 打开shape文件
	feaset := shape.OpenShape(filename, true, []string{})
	// // 创建地图
	gmap := mapping.NewMap(draw.Default)
	gmap.AddFeatureLayer(feaset, nil)
	return gmap
}

func testDrawMap() {
	gmap := mapping.NewMap(draw.Default)
	mapname := gPath + gTitle + "." + base.EXT_MAP_FILE

	gmap.Open(mapname) // chinapnt_84 JBNTBHTB
	// gmap.Layers[0].Style.LineColor = color.RGBA{255, 0, 0, 255}

	tr := base.NewTimeRecorder()

	gmap.Prepare(1024, 768)
	// gmap.Zoom(2)
	gmap.Draw()

	// 输出图片文件
	tr.Output("draw map")
	gmap.Output2File(gPath+gTitle+".jpg", "jpg")
	tr.Output("save picture file")
	// debug.SetGCPercent(1)
	// tr.Output("SetGCPercent")

	// fmt.Println("")
	// time.Sleep(10000000000)
	fmt.Println("DONE")
}

func testMapFile() {
	tr := base.NewTimeRecorder()

	gmap := startMap()
	// gmap.Layers[0].Style.FillColor = color.RGBA{25, 200, 20, 255}
	// gmap.Layers[0].Style.LineColor = color.RGBA{225, 20, 20, 255}
	// 设置位图大小
	gmap.Prepare(1024, 768)

	// gmap.Zoom(5)
	// 绘制
	gmap.Draw()
	// // 输出图片文件
	gmap.Output2File(gPath+gTitle+".png", "png")

	mapfile := gPath + gTitle + "." + base.EXT_MAP_FILE

	gmap.Save(mapfile)

	nmap := mapping.NewMap(draw.Default)
	nmap.Open(mapfile)
	nmap.Prepare(1024, 768)
	nmap.Draw()
	// // 输出图片文件
	nmap.Output2File(gPath+gTitle+"2.png", "png")

	// // 记录时间
	tr.Output("testMapFile total")
}

func testLine2Point() {
	tr := base.NewTimeRecorder()

	title := "china2" // JBNTBHTB chinapnt_84

	fromParams := data.NewConnParams()
	fromParams["filename"] = "c:/temp/" + title + ".udbx" // line.udbx
	fromParams["type"] = string(data.StoreSqlite)
	fromStore := data.NewDatastore(data.StoreSqlite)
	fromStore.Open(fromParams)
	fromFeaset := fromStore.GetFeasetByNum(0)
	fromFeaset.Open()

	toParams := data.NewConnParams()
	toParams["filename"] = "c:/temp/" + "railway" + ".sqlite"
	toParams["type"] = string(data.StoreSqlite)
	toStore := data.NewDatastore(data.StoreSqlite)
	toStore.Open(toParams)

	var cvt data.Converter
	cvt.Polyline2Point(fromFeaset, toStore, "point")
	tr.Output("Polyline2Point")
}

func testDrawTiff() {
	tr := base.NewTimeRecorder()

	// title := "raster" // JBNTBHTB chinapnt_84
	// filename := "C:/BigData/10_Data/testimage/image/filelist.txt"
	// filename := "C:/temp/filelist.gmr"
	// filename := "C:/temp/raster.txt"
	path := "C:/BigData/10_Data/images/imagebig2/"
	// path := "C:/BigData/10_Data/testimage/image2/"
	filename := path + "image2."

	var raset data.MosaicRaset
	raset.Open(filename + "gmr")
	// raset.Save(filename + "gmr")
	tr.Output("open data")

	gmap := mapping.NewMap(draw.Default)
	gmap.AddRasterLayer(&raset)
	gmap.Prepare(1024, 768)
	gmap.Draw()

	// 输出图片文件
	tr.Output("draw map")
	gmap.Output2File(gPath+"image.jpg", "jpg")
	tr.Output("save picture file")

	gmap.Save(path + "image2.gmp")

	amap := mapping.NewMap(draw.Default)
	amap.Open(path + "image2.gmp")
	amap.Prepare(1024, 768)
	amap.Draw()
	amap.Output2File(gPath+"image2.jpg", "jpg")

}

func buildMosaic() {
	tr := base.NewTimeRecorder()

	path := "C:/BigData/10_Data/images/imagebig2/"
	// path := "C:/BigData/10_Data/testimage/image2/"
	var raset data.MosaicRaset
	raset.Build(path, path+"image2.gmr")
	tr.Output("build")

	// raset.Open(path + "image2.gmr")
	gmap := mapping.NewMap(draw.Default)
	gmap.AddRasterLayer(&raset)
	gmap.Prepare(1024, 768)
	gmap.Draw()

	// 输出图片文件
	tr.Output("draw map")
	gmap.Output2File(gPath+"image.jpg", "jpg")
	gmap.Save(path + "image2.gmp")

	tr.Output("save picture file")
}
