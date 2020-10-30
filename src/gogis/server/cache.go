package server

import (
	"fmt"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"time"

	"gogis/data"
	"gogis/mapping"

	"github.com/drone/routes"
)

var gMap = mapping.NewMap()

var gPath = "c:/temp/"

var gEpsg = mapping.Epsg4326

func startMap() {

	feaset := data.OpenShape(gPath + "JBNTBHTB.shp")

	// shp := new(data.ShapeStore)
	// params := data.NewCoonParams()
	// params["filename"] = gPath + "JBNTBHTB.shp"
	// shp.Open(params)

	// // 创建地图
	// gmap := mapping.NewMap()
	// feaset, _ := shp.GetFeasetByNum(0)
	gMap.AddLayer(feaset)
	gMap.RebuildBBox()
}

func StartServer() {
	fmt.Println("正在启动WEB服务...")
	startTime := time.Now().UnixNano()

	mux := routes.New()
	mux.Get("/:level/:col/:row", getTile)
	// mux.Get("/:map/:size/:row/:col/:r", getmap)
	http.Handle("/", mux)

	// gMap.Open(gPath + "JBNTBHTB.txt")
	startMap()

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Println("WEB服务启动完毕，花费时间：...", seconds)

	http.ListenAndServe(":8088", nil)
	fmt.Println("服务已停止")
}

func getTile(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now().UnixNano()

	params := r.URL.Query()
	// mapname := params.Get(":map")
	level, _ := strconv.Atoi(params.Get(":level"))
	col, _ := strconv.Atoi(params.Get(":col"))
	row, _ := strconv.Atoi(params.Get(":row"))

	if !gMap.BBox.IsIntersect(mapping.CalcBBox(level, col, row, gEpsg)) {
		return
	}

	fmt.Println("get cache,", "level=", level, "col=", col, "row=", row)
	cachefile := getFileName(level, col, row)
	data, exist := getCache(cachefile)
	if exist {
		w.Write(data)
		// fmt.Println("map cache:", data)
	} else {

		// 这里根据地图名字，输出并返回图片
		// gmap, ok := gmaps[mapname]
		// if ok {
		mapTile := mapping.NewMapTile(gMap, gEpsg)

		tmap := mapTile.CacheOneTile2Map(level, col, row, nil)
		if tmap != nil {
			png.Encode(w, tmap.OutputImage())

			// 缓存起来
			os.MkdirAll(getPath(level, col), os.ModePerm)
			f, err := os.Create(cachefile)
			if err != nil {
				fmt.Println("create file error: ", err)
			}
			fmt.Println("create cache file: ", cachefile)
			err = png.Encode(f, tmap.OutputImage())
			if err != nil {
				fmt.Println("Encode error: ", err)
			}
			f.Close()
		}
		// } else {
		// 	fmt.Fprintf(w, "cannot find map %s", mapname)
		// }
	}

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Println("time: ", seconds, "毫秒")
}

func getPath(level int, col int) string {
	path := gPath + "cache/"
	path += strconv.Itoa(level) + "/" + strconv.Itoa(col) + "/"
	return path
}

func getFileName(level int, col int, row int) string {
	return getPath(level, col) + strconv.Itoa(row) + ".png"
}

func getCache(filename string) (data []byte, exist bool) {
	f, err := os.Open(filename)
	defer f.Close()

	if err != nil && os.IsNotExist(err) {
		fmt.Println("file: ", filename, " not exist!")
		exist = false
	} else {
		fmt.Println("file: ", filename, " is finded!")
		info, _ := os.Stat(filename)
		data = make([]byte, info.Size())
		f.Read(data)
		exist = true
	}
	return
}

// var gmaps map[string]*gogis.Map
