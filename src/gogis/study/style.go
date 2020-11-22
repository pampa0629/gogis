package main

import (
	"encoding/json"
	"fmt"
	"gogis/data"
	"gogis/mapping"
	"image/color"
	"time"
)

func stylemain() {
	// testColor()
	stestMapFile()
	return

}

type TestColor struct {
	color color.RGBA
}

func testColor() {
	var tc, tc2, tc3 TestColor
	tc.color = color.RGBA{123, 33, 44, 55}
	data, _ := json.Marshal(tc)
	json.Unmarshal(data, &tc2)
	json.Unmarshal(data, tc3)
	fmt.Println(string(data), tc, tc2, tc3)

}

func stestMapFile() {
	var gPath = "C:/temp/"

	var gTitle = "chinapnt_84"

	// var gTitle = "DLTB"
	// var gTitle = "JBNTBHTB"

	var gExt = ".shp"

	var filename = gPath + gTitle + gExt

	feaset := data.OpenShape(filename)
	gmap := mapping.NewMap()
	gmap.AddLayer(feaset)
	fmt.Println("map:", gmap)

	gmap.Prepare(1024, 768)

	// gmap.Zoom(5)
	// 绘制
	gmap.Draw()
	// // 输出图片文件
	gmap.Output2File(gPath+gTitle+".png", "png")

	mapfile := gPath + gTitle + ".map"

	gmap.Save(mapfile)
	// ============================

	startTime := time.Now().UnixNano()
	nmap := mapping.NewMap()
	nmap.Open(mapfile)
	nmap.Prepare(1024, 768)
	nmap.Draw()
	// // 输出图片文件
	nmap.Output2File(gPath+gTitle+"2.png", "png")

	// // 记录时间
	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("time: %f 毫秒", seconds)
}
