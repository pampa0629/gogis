// gogis rest服务总入口

package server

import (
	"fmt"
	"net/http"
	"strconv"

	"gogis/base"
	"gogis/mapping"

	"github.com/drone/routes"
)

var mapServices MapServices

func StartServer(gmsfile string, cachepath string) {
	fmt.Println("正在启动WEB服务...")
	tr := base.NewTimeRecorder()

	mux := routes.New()
	mux.Get("/:map/:level/:col/:row", getTile)
	mux.Get("/:mapserver", getMapNames)
	http.Handle("/", mux)

	// 读取地图服务配置文件
	mapServices.Open(gmsfile, cachepath)

	tr.Output("start web server finished,")

	http.ListenAndServe(":8088", nil)
	defer mapServices.Close()
	fmt.Println("服务已停止")
}

// 得到已发布的地图服务列表
func getMapNames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	params := r.URL.Query()
	params.Get(":mapserver")
	data := make([]byte, 0)

	count := len(mapServices.mapServices)
	i := 0
	for t, _ := range mapServices.mapServices {
		data = append(data, []byte(t)...)
		// 最后一个不加 ;
		if i != count-1 {
			data = append(data, []byte(";")...)
		}
		i++
	}
	fmt.Println("getMapNames:", data)
	w.Write(data)
}

// 通过地图名、层级和行列号，得到对应的地图瓦片
// 有缓存就用缓存，没有就现生成，并缓存起来
func getTile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // 跨域

	params := r.URL.Query()
	mapname := params.Get(":map")
	level, _ := strconv.Atoi(params.Get(":level"))
	col, _ := strconv.Atoi(params.Get(":col"))
	row, _ := strconv.Atoi(params.Get(":row"))

	tr := base.NewTimeRecorder()
	data := mapServices.GetTile(mapname, level, col, row, mapping.Epsg4326)
	if data != nil && len(data) > 0 {
		w.Write(data)
		tr.Output("get tile")
	}

	return
}
