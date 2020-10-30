package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"gogis/data"
	"gogis/mapping"
	"gogis/server"

	"time"

	"github.com/chai2010/tiff"
	"github.com/lukeroth/gdal"
	// dbf "github.com/SebastiaanKlippert/go-foxpro-dbf"
)

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

// func testRest2() {
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

func testRest() {
	server.StartServer()

	// gmaps = make(map[string]*mapping.Map)
	// // 这里加载所有地图
	// names := []string{"c:/temp/DLTB.txt", "c:/temp/australia.txt", "c:/temp/JBNTBHTB.txt"}
	// for _, name := range names {
	// 	gmap := mapping.NewMap()
	// 	gmap.Open(name)
	// 	gmap.Prepare(1024, 1024) // todo ...
	// 	title := gogis.GetTile(name)
	// 	gmaps[title] = gmap
	// }

}

func testGdal() {
	startTime := time.Now().UnixNano()

	var wg *sync.WaitGroup = new(sync.WaitGroup)

	for i := 0; i < 1; i++ {
		index := strconv.Itoa(i)
		wg.Add(1)
		go testOneTiff(index, wg)
	}

	wg.Wait()

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Println("time: ", seconds, "毫秒")
}

func drawTiff() {
	startTime := time.Now().UnixNano()

	filename := "C:\\temp\\A49C001003-0.tiff"
	dt, _ := gdal.Open(filename, 0)
	band := dt.RasterBand(1)
	band2 := dt.RasterBand(2)
	band3 := dt.RasterBand(3)
	sx := band.XSize()
	sy := band.YSize()
	bx, by := band.BlockSize()
	fmt.Println("size: ", sx, sy, bx, by)
	rdt := band.RasterDataType()
	fmt.Println("data type: ", rdt)
	// band.b
	// data := make([][]byte, sy)

	ix, iy := 1024, 768
	img := image.NewNRGBA(image.Rect(0, 0, ix, iy))
	rx := float64(ix) / float64(sx)
	ry := float64(iy) / float64(sy)
	fmt.Println("rest  ratio: ", rx, ry)

	for i := 0; i < sy; i++ {
		data := make([]uint8, sx)
		data2 := make([]uint8, sx)
		data3 := make([]uint8, sx)
		band.ReadBlock(0, i, unsafe.Pointer(&data[0]))
		band2.ReadBlock(0, i, unsafe.Pointer(&data2[0]))
		band3.ReadBlock(0, i, unsafe.Pointer(&data3[0]))
		for j := 0; j < sx; j++ {
			x, y := int(float64(i)*rx), int(float64(j)*ry)
			img.Set(x, y, color.RGBA{data[j], data2[j], data3[j], 255})
		}
	}
	imgfile, _ := os.Create("c:/temp/image.jpeg")
	jpeg.Encode(imgfile, img, nil)

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Println("time: ", seconds, "毫秒")
}

func testOneTiff(index string, wg *sync.WaitGroup) {
	defer wg.Done()

	filename := "C:\\temp\\A49C001003-" + index + ".tiff"
	fmt.Println("filename: ", filename)

	dt, _ := gdal.Open(filename, 0)
	// dt.BuildOverviews()
	fmt.Println("file list: ", dt.FileList())
	fmt.Println("GeoTransform: ", dt.GeoTransform())
	fmt.Println("Projection: ", dt.Projection())

	rc := dt.RasterCount()
	fmt.Println("Count: ", rc)

	band := dt.RasterBand(1)
	xs := band.XSize()
	ys := band.YSize()
	x, y := band.BlockSize()
	fmt.Println("size: ", xs, ys, x, y)
	// var data [10508 * 7028 * 1]byte

	// data := make([]byte, x*y*64)
	// gdal.Warp()
	rdt := band.RasterDataType()
	fmt.Println("data type: ", rdt)
	// band.b
	data := make([][]byte, 7028)

	sum := 0
	count := 0
	for i := 0; i < ys; i++ {
		// go func(band gdal.RasterBand, data [][]byte, i int) {
		data[i] = make([]byte, 10508)
		band.ReadBlock(0, i, unsafe.Pointer(&data[i][0]))
		// for _, v := range data[i] {
		// 	sum += int(v)
		// if v != 0 {
		// 	// fmt.Println("value: ", i, v)
		// 	count++
		// }
		// }
		// }(band, data, i)
	}

	fmt.Println("sum of data: ", sum)
	fmt.Println("count: ", count)

	// band.ReadBlock(0,0,)
	fmt.Println("BandNumber: ", band.BandNumber())
	fmt.Println("OverviewCount: ", band.OverviewCount())
	for i := 0; i < band.OverviewCount(); i++ {
		ov := band.Overview(i)
		x, y := ov.BlockSize()
		fmt.Println("size: ", i, x, y)

	}

}

func testTiff() {
	tiffname := "C:\\temp\\A49C001003.tiff"

	data, _ := ioutil.ReadFile(tiffname)
	fmt.Println(len(data))

	// Decode tiff
	img, err := tiff.Decode(bytes.NewReader(data))

	fmt.Println(err)
	fmt.Println(img.Bounds(), img.ColorModel())
	// fmt.Println(img)
	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()
	// add := uint32(0)
	for i := 0; i < dy; i++ {
		for j := 0; j < dx; j++ {
			clr := img.At(i, j)
			r, g, b, a := clr.RGBA()
			one := r + g + b
			if one != 0 {
				fmt.Println(i, j, one, clr, r, g, b, a)
			}
			// add += r
			// add += g
			// add += b
			// add += a
		}
	}

}

func testMap() {
	gmap := mapping.NewMap()
	gmap.Open("c:/temp/JBNTBHTB.txt")

	startTime := time.Now().UnixNano()
	gmap.Prepare(4000, 3000)
	gmap.Draw()

	// 输出图片文件
	gmap.Output2File("c:/temp/JBNTBHTB.png", "png")
	// gmap.Resize(3000, 4000)
	// gmap.Draw()

	// // 输出图片文件
	// gmap.Output("c:/temp/result2.png")
	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("time: %f 毫秒", seconds)
}

func testCache() {
	// names := []string{"c:/temp/australia.txt", "c:/temp/DLTB.txt", "c:/temp/JBNTBHTB.txt"}

	// gmap := gogis.NewMap()
	// gmap.Open(names[2])
	startTime := time.Now().UnixNano()

	gmap := startMap()
	mapTile := mapping.NewMapTile(gmap, mapping.Epsg4326)
	mapTile.Cache("c:/temp/cache/")
	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e9)
	fmt.Printf("time: %f 秒", seconds)
}

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
	params := data.NewConnParams()
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

// var filename = "C:/temp/chinapnt_84.shp"

var filename = "C:/temp/DLTB.shp"

// var filename = "C:/temp/JBNTBHTB.shp"

func main() {
	// testRest()
	testMapFile()
	// testMap()
	// testCache()
	// testVecPyramid()

	// testTiff()
	// testGdal()
	// drawTiff()
	// test()
	// testDBF()
	// testSQL()
	return

	startTime := time.Now().UnixNano()

	// 打开shape文件
	shp := new(data.ShapeStore)
	// var params data.ConnParams
	params := data.NewConnParams()
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
	gmap.Output2File("c:/temp/result2.png", "png")

	// // gmap.Save("c:/temp/map.txt")

	// // 记录时间
	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("time: %f 毫秒", seconds)
}

func startMap() *mapping.Map {
	// 打开shape文件
	feaset := data.OpenShape(filename)

	// shp := new(data.ShapeStore)
	// // var params data.ConnParams
	// params := data.NewCoonParams()
	// params["filename"] = filename
	// // params = make(map[string]string)

	// shp.Open(params)

	// // 创建地图
	gmap := mapping.NewMap()
	// feaset, _ := shp.GetFeasetByNum(0)
	gmap.AddLayer(feaset)
	return gmap
}

func testMapFile() {
	startTime := time.Now().UnixNano()

	gmap := startMap()
	// 设置位图大小
	gmap.Prepare(1024, 768)

	// gmap.Zoom(5)
	// 绘制
	gmap.Draw()
	// // 输出图片文件
	gmap.Output2File("c:/temp/result.png", "png")

	mapfile := "c:/temp/map.txt"

	gmap.Save(mapfile)

	nmap := mapping.NewMap()
	nmap.Open(mapfile)
	nmap.Prepare(1024, 768)
	nmap.Draw()
	// // 输出图片文件
	nmap.Output2File("c:/temp/result2.png", "png")

	// // 记录时间
	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("time: %f 毫秒", seconds)
}
