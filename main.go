package main

import (
	"fmt"
	"image/color"
	"os"

	"gogis/base"
	"gogis/data"
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
	testRest()

	// testMapFile()
	// testDrawMap()
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

	gmap.Prepare(256, 256)
	gmap.Zoom(10)
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
