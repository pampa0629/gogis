package main

import (
	"fmt"
	"gogis/data"
	"strings"
	"time"
)

// 查询
func testQuery() {
	var gPath = "C:/temp/"

	// var gTitle = "chinapnt_84"

	var gTitle = "DLTB"

	// var gTitle = "JBNTBHTB"

	var gExt = ".shp"

	var filename = gPath + gTitle + gExt

	startTime := time.Now().UnixNano()

	shp := new(data.ShapeStore)
	params := data.NewConnParams()
	params["filename"] = filename

	shp.Open(params)
	// set, _ := shp.GetFeasetByNum(0)

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("open time: %f 毫秒", seconds)
	startTime = time.Now().UnixNano()

	var def data.QueryDef
	def.Fields = []string{"TKXS"}
	def.Where = "TKXS>10"

	// ft := set.QueryByDef(def)
	// fmt.Println("fea count:", ft.Count())

	for {
		// _, ok := ft.Next()
		// if !ok {
		// 	break
		// }
		// fmt.Println(fea)
	}

	endTime = time.Now().UnixNano()
	seconds = float64((endTime - startTime) / 1e6)
	fmt.Printf("query time: %f 毫秒", seconds)
}

func testDBF() {
	startTime := time.Now().UnixNano()

	var gPath = "C:/temp/"

	// var gTitle = "chinapnt_84"

	var gTitle = "DLTB"

	// var gTitle = "JBNTBHTB"

	var gExt = ".shp"

	var filename = gPath + gTitle + gExt

	var dbf data.DbfFile
	// dbfName := "c:/temp/Provinces.dbf"
	dbfName := strings.TrimSuffix(filename, ".shp") + ".dbf"
	dbf.Open(dbfName, "GB18030")

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("time: %f 毫秒", seconds)
}
