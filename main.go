package main

import (
	"fmt"
	"strings"

	"gogis/data"
	"gogis/mapping"

	"time"
	// dbf "github.com/SebastiaanKlippert/go-foxpro-dbf"
)

// func cachefile(mapname string, size int, row int, col int, ratio int) string {
// 	file := "c:/temp/cache/"
// 	file += mapname + strconv.Itoa(size) + strconv.Itoa(row) + strconv.Itoa(col) + strconv.Itoa(ratio)
// 	return file
// }

// func getcache(filename string) (data []byte, exist bool) {
// 	f, err := os.Open(filename)
// 	if err != nil && os.IsNotExist(err) {
// 		fmt.Printf("file not exist!\n")
// 		exist = false
// 	} else {
// 		fmt.Printf("file exist!\n")
// 		info, _ := os.Stat(filename)
// 		data = make([]byte, info.Size())
// 		f.Read(data)
// 		exist = true
// 	}

// 	defer f.Close()
// 	return
// }

// func getmap(w http.ResponseWriter, r *http.Request) {
// 	startTime := time.Now().UnixNano()

// 	params := r.URL.Query()
// 	mapname := params.Get(":map")
// 	size, _ := strconv.Atoi(params.Get(":size"))
// 	row, _ := strconv.Atoi(params.Get(":row"))
// 	col, _ := strconv.Atoi(params.Get(":col"))
// 	ratio, _ := strconv.Atoi(params.Get(":r"))
// 	fmt.Println("get map", mapname, size, row, col, ratio)
// 	cachefile := cachefile(mapname, size, row, col, ratio)
// 	data, exist := getcache(cachefile)
// 	if exist {
// 		w.Write(data)
// 		// fmt.Println("map cache:", data)
// 	} else {
// 		// 这里根据地图名字，输出并返回图片
// 		gmap, ok := gmaps[mapname]
// 		if ok {
// 			nmap := gmap.Copy()
// 			nmap.Zoom(float64(ratio))

// 			gogis.SetParams(gmap, nmap, size, row, col)
// 			nmap.Prepare(size, size)

// 			nmap.Draw()
// 			png.Encode(w, nmap.OutputImage())

// 			// 缓存起来
// 			f, _ := os.Create(cachefile)
// 			defer f.Close()
// 			png.Encode(f, nmap.OutputImage())
// 		} else {
// 			fmt.Fprintf(w, "cannot find map %s", mapname)
// 		}
// 	}

// 	endTime := time.Now().UnixNano()
// 	seconds := float64((endTime - startTime) / 1e6)
// 	fmt.Println("time: ", seconds, "毫秒")
// }

// var gmaps map[string]*gogis.Map

// func testRest() {
// 	fmt.Println("正在启动WEB服务...")
// 	startTime := time.Now().UnixNano()

// 	gmaps = make(map[string]*gogis.Map)
// 	// 这里加载所有地图
// 	names := []string{"c:/temp/DLTB.txt", "c:/temp/australia.txt", "c:/temp/JBNTBHTB.txt"}
// 	for _, name := range names {
// 		gmap := gogis.NewMap()
// 		gmap.Open(name)
// 		gmap.Prepare(1024, 1024) // todo ...
// 		title := gogis.GetTile(name)
// 		gmaps[title] = gmap
// 	}

// 	mux := routes.New()
// 	mux.Get("/:map/:size/:row/:col/:r", getmap)
// 	http.Handle("/", mux)

// 	endTime := time.Now().UnixNano()
// 	seconds := float64((endTime - startTime) / 1e6)
// 	fmt.Println("WEB服务启动完毕，花费时间：...", seconds)

// 	http.ListenAndServe(":8088", nil)
// 	fmt.Println("服务已停止")
// }

// func testTiff() {
// 	tiffname := "C:\\temp\\5.tiff"

// 	data, _ := ioutil.ReadFile(tiffname)
// 	fmt.Println(len(data))

// 	// Decode tiff
// 	_, err := tiff.Decode(bytes.NewReader(data))

// 	fmt.Println(err)
// 	// fmt.Println(img)
// }

// func testMap() {
// 	gmap := gogis.NewMap()
// 	gmap.Open("c:/temp/JBNTBHTB.txt")

// 	startTime := time.Now().UnixNano()
// 	gmap.Prepare(4000, 3000)
// 	gmap.Draw()

// 	// 输出图片文件
// 	gmap.Output("c:/temp/JBNTBHTB.png", "png")
// 	// gmap.Resize(3000, 4000)
// 	// gmap.Draw()

// 	// // 输出图片文件
// 	// gmap.Output("c:/temp/result2.png")
// 	endTime := time.Now().UnixNano()
// 	seconds := float64((endTime - startTime) / 1e6)
// 	fmt.Printf("time: %f 毫秒", seconds)
// }

// func testCache() {
// 	names := []string{"c:/temp/australia.txt", "c:/temp/DLTB.txt", "c:/temp/JBNTBHTB.txt"}

// 	gmap := gogis.NewMap()
// 	gmap.Open(names[2])
// 	startTime := time.Now().UnixNano()
// 	gmap.Cache("c:/temp/cache2/")
// 	endTime := time.Now().UnixNano()
// 	seconds := float64((endTime - startTime) / 1e9)
// 	fmt.Printf("time: %f 秒", seconds)
// }

// var filename = "C:/temp/chinapnt_84.shp"

// var filename = "C:/temp/DLTB.shp"

var filename = "C:/temp/JBNTBHTB.shp"

// func testVecPyramid() {
// 	shp := new(gogis.ShapeFile)

// 	shp.Open(filename)
// 	shp.Load()

// 	startTime := time.Now().UnixNano()
// 	shp.BuildVecPyramid()

// 	// 记录时间
// 	endTime := time.Now().UnixNano()
// 	seconds := float64((endTime - startTime) / 1e6)
// 	fmt.Printf("time: %f 毫秒", seconds)
// }

// 查询
func testQuery() {
	startTime := time.Now().UnixNano()

	shp := new(data.ShapeStore)
	params := data.NewCoonParams()
	params["filename"] = filename

	shp.Open(params)
	set, _ := shp.GetFeasetByNum(0)

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("open time: %f 毫秒", seconds)
	startTime = time.Now().UnixNano()

	var def data.QueryDef
	def.Fields = []string{"TKXS"}
	def.Wheres = []string{"TKXS>10"}

	ft := set.QueryByDef(def)
	fmt.Println("fea count:", ft.Count())

	for {
		_, ok := ft.Next()
		if !ok {
			break
		}
		// fmt.Println(fea)
	}

	endTime = time.Now().UnixNano()
	seconds = float64((endTime - startTime) / 1e6)
	fmt.Printf("query time: %f 毫秒", seconds)
}

func testDBF() {
	startTime := time.Now().UnixNano()

	var dbf data.DbfFile
	// dbfName := "c:/temp/Provinces.dbf"
	dbfName := strings.TrimSuffix(filename, ".shp") + ".dbf"
	dbf.Open(dbfName, "GB18030")

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("time: %f 毫秒", seconds)
}

func main() {
	// testRest()
	// testMap()
	// testCache()
	// testVecPyramid()

	// testTiff()
	// test()
	testDBF()
	// testSQL()
	return

	startTime := time.Now().UnixNano()

	// 打开shape文件
	shp := new(data.ShapeStore)
	// var params data.ConnParams
	params := data.NewCoonParams()
	params["filename"] = filename
	// params = make(map[string]string)

	shp.Open(params)

	// // 创建地图
	gmap := mapping.NewMap()
	feaset, _ := shp.GetFeasetByNum(0)
	gmap.AddLayer(feaset)
	// 设置位图大小
	gmap.Prepare(1024, 768)

	// gmap.Zoom(5)
	// 绘制
	gmap.Draw()
	// // 输出图片文件
	gmap.Output("c:/temp/result2.png", "png")

	// // gmap.Save("c:/temp/map.txt")

	// // 记录时间
	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("time: %f 毫秒", seconds)
}

func test() {

}
