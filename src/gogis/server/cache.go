package server

import (
	"fmt"
	"image/png"
	"net/http"
	"os"
	"strconv"

	"gogis/base"
	"gogis/mapping"

	"github.com/drone/routes"
)

// var gMap = mapping.NewMap()

var gPath = "c:/temp/"

// var gTitles = []string{"DLTB"}

var gTitles = []string{"JBNTBHTB", "chinapnt_84", "DLTB"}

var gMaps map[string]*mapping.Map

var gEpsg = mapping.Epsg4326

func startMap() {
	fmt.Println("gTitles:", gTitles)
	gMaps = make(map[string]*mapping.Map, len(gTitles))
	for _, title := range gTitles {
		gMaps[title] = mapping.NewMap()
		mapfile := gPath + title + ".map"
		fmt.Println("open map:", mapfile)
		gMaps[title].Open(mapfile)
	}
}

func StartServer() {
	fmt.Println("正在启动WEB服务...")
	tr := base.NewTimeRecorder()

	mux := routes.New()
	mux.Get("/:map/:level/:col/:row", getTile)
	// mux.Get("/:map/:size/:row/:col/:r", getmap)
	http.Handle("/", mux)

	// gMap.Open(gPath + "JBNTBHTB.txt")
	startMap()

	tr.Output("start web server finished,")

	http.ListenAndServe(":8088", nil)
	fmt.Println("服务已停止")
}

func getTile(w http.ResponseWriter, r *http.Request) {
	tr := base.NewTimeRecorder()

	params := r.URL.Query()
	mapname := params.Get(":map")
	level, _ := strconv.Atoi(params.Get(":level"))
	col, _ := strconv.Atoi(params.Get(":col"))
	row, _ := strconv.Atoi(params.Get(":row"))

	gMap := gMaps[mapname]

	// fmt.Println("BBox:", gMap.BBox)
	bbox := mapping.CalcBBox(level, col, row, gEpsg)
	fmt.Println("map:", mapname, "level=", level, "col=", col, "row=", row, "BBox:", bbox)

	if !gMap.BBox.IsIntersect(bbox) {
		return
	}

	cachefile := getFileName(mapname, level, col, row)
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
			os.MkdirAll(getPath(mapname, level, col), os.ModePerm)
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

	tr.Output("get tile")
}

func getPath(mapname string, level int, col int) string {
	path := gPath + "cache/" + mapname + "/"
	path += strconv.Itoa(level) + "/" + strconv.Itoa(col) + "/"
	return path
}

func getFileName(mapname string, level int, col int, row int) string {
	return getPath(mapname, level, col) + strconv.Itoa(row) + ".png"
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
