package main

import (
	"fmt"
	"image/color"
	"os"

	"gogis/base"
	"gogis/data"
	"gogis/index"
	"gogis/mapping"
	"gogis/server"
)

func testRest() {
	rootpath := os.Args[0]
	datapath := "./data/"
	cachepath := "./cache/"
	if len(os.Args) >= 3 {
		datapath = os.Args[1]
		cachepath = os.Args[2]
	}
	datapath = base.GetAbsolutePath(rootpath, datapath)
	cachepath = base.GetAbsolutePath(rootpath, cachepath)
	fmt.Println("app path:", rootpath, ", data path:", datapath, ", cache path:", cachepath)
	server.StartServer(datapath+"/gogis.gms", cachepath)
}

var gPath = "C:/temp/"

// var gTitle = "chinapnt_84"

var gTitle = "DLTB"

// var gTitle = "JBNTBHTB"

var gExt = ".shp"

var filename = gPath + gTitle + gExt

func main() {
	// testRest()

	testDrawMap()
	// testMapTile()
	// testIndex()

	// testSqliteMap()
	// testHbaseMap()
	fmt.Println("DONE!")
	return
}

func testIndex() {
	// codes := []int{0, 3, 16, 65, 66, 264, 267, 1058, 1060, 1069, 1070, 1071, 1072}
	// codes := []int{0, 4, 17, 69, 277, 278, 1112, 1115}
	temp := index.LoadGix(gPath + gTitle + ".gix")
	idx, _ := temp.(*index.ZOrderIndex)

	code2ids := idx.Data()
	// println("ids:", co)
	// ids := ""
	for _, v := range code2ids {
		for _, vv := range v {
			if vv.Id == 559055 {
				fmt.Println("bbox:", vv.Bbox)
				code := idx.GetCode(vv.Bbox)
				fmt.Println("code:", code)
			}
		}
	}

	// count := 0
	// for _, v := range codes {
	// 	println("code:", v, "id count:", len(code2ids[v]))
	// 	count += len(code2ids[v])
	// }
	// println("tatol count:", count)
}

func testHbaseMap() {
	tr := base.NewTimeRecorder()
	var store data.HbaseStore
	params := data.NewConnParams()
	params["address"] = "localhost:2181"
	store.Open(params)
	feaset, _ := store.GetFeasetByName(gTitle)
	feaset.Open()
	tr.Output("open hbase db")

	gmap := mapping.NewMap()
	gmap.AddLayer(feaset)
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
	var sqlDB data.SqliteStore
	params := data.NewConnParams()
	params["filename"] = "C:/temp/JBNTBHTB.sqlite"
	sqlDB.Open(params)
	feaset, _ := sqlDB.GetFeasetByNum(0)
	feaset.Open()
	tr.Output("open sqlite db")

	gmap := mapping.NewMap()
	gmap.AddLayer(feaset)
	gmap.Prepare(1024, 768)
	gmap.Zoom(2)
	gmap.Draw()
	// 输出图片文件
	gmap.Output2File("C:/temp/JBNTBHTB.jpg", "jpg")

	tr.Output("draw sqlite map")
	fmt.Println("DONE!")
}

func testMapTile() {
	tr := base.NewTimeRecorder()

	gmap := mapping.NewMap()
	gmap.Open("c:/temp/JBNTBHTB-hbase.gmp") //sqlite hbase
	maptile := mapping.NewMapTile(gmap, mapping.Epsg4326)
	// this.tilestore = new(data.LeveldbTileStore) // data.FileTileStore LeveldbTileStore
	// this.tilestore.Open(path, mapname)
	tmap, _ := maptile.CacheOneTile2Map(6, 101, 23, nil)
	tilename := gPath + gTitle + ".jpg"
	fmt.Println(tilename)
	tmap.Output2File(tilename, "jpg")

	return

	idx := index.LoadGix(gPath + gTitle + ".gix")
	idboxs := idx.(*index.ZOrderIndex).Data()
	ids1 := getIds(tmap.BBox, idboxs)
	fmt.Println(len(ids1))

	ids2 := idx.Query(tmap.BBox)
	fmt.Println(len(ids2))
	codes := idx.(*index.ZOrderIndex).QueryDB(tmap.BBox)
	fmt.Println("codes:", codes)
	// index.Bits2code()
	return

	mapids := make(map[int64]byte)
	for _, v := range ids2 {
		mapids[v] = 1
	}
	for _, v := range ids1 {
		count := len(mapids)
		mapids[v] = 1
		if len(mapids) > count {
			code, bbox := id2code(v, idboxs)
			bits := index.Code2bits(code)
			fmt.Println("id:", v, "code:", code, "bits:", bits, "bbox:", bbox)
		}
	}

	tr.Output("map tile")
}

func id2code(id int64, idboxs [][]index.Idbbox) (code int, bbox base.Rect2D) {
	for i, v := range idboxs {
		for _, vv := range v {
			if vv.Id == id {
				code = i
				bbox = vv.Bbox
				return
			}
		}
	}
	return
}

func getIds(bbox base.Rect2D, idboxs [][]index.Idbbox) (ids []int64) {
	for _, v := range idboxs {
		for _, vv := range v {
			// if vv.Bbox.IsIntersect(bbox) {
			ids = append(ids, vv.Id)
			// }
		}
	}
	return
}

func startMap() *mapping.Map {
	// 打开shape文件
	feaset := data.OpenShape(filename)
	// // 创建地图
	gmap := mapping.NewMap()
	gmap.AddLayer(feaset)
	return gmap
}

func testDrawMap() {
	gmap := mapping.NewMap()
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
	gmap.Layers[0].Style.FillColor = color.RGBA{25, 200, 20, 255}
	gmap.Layers[0].Style.LineColor = color.RGBA{225, 20, 20, 255}
	// 设置位图大小
	gmap.Prepare(1024, 768)

	// gmap.Zoom(5)
	// 绘制
	gmap.Draw()
	// // 输出图片文件
	gmap.Output2File(gPath+gTitle+".png", "png")

	mapfile := gPath + gTitle + "." + base.EXT_MAP_FILE

	gmap.Save(mapfile)

	nmap := mapping.NewMap()
	nmap.Open(mapfile)
	nmap.Prepare(1024, 768)
	nmap.Draw()
	// // 输出图片文件
	nmap.Output2File(gPath+gTitle+"2.png", "png")

	// // 记录时间
	tr.Output("testMapFile total")
}
