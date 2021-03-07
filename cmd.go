package main

import (
	"flag"
	"fmt"
	"os"

	"gogis/base"
	_ "gogis/data/memory"
	"gogis/desktop"
	"gogis/draw"
	"gogis/mapping"
	"gogis/server"
)

func main() {
	fmt.Println(os.Args)
	if len(os.Args) < 2 {
		fmt.Println(`input "gogis help" to know how to use.`)
		other()
		return
	}

	subCommand := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...)

	gmsfile := flag.String("gmsfile", "", "gogis map service file, ext is *.gms")
	cachepath := flag.String("cachepath", "./cache/", "the path of saving cache files")
	mapfile := flag.String("mapfile", "", "gogis map doc file, ext is *.gmp")
	maptype := flag.String("maptype", "png", "such as: mvt/png/jpg/webp")
	picfile := flag.String("picfile", "png", "the picture file name of map drawn")
	width := flag.Int("width", 1024, "the width of desktop or picfile")
	height := flag.Int("height", 768, "the height of desktop or picfile")
	batch := flag.Int("batch", 0, "the obj count of one batch when drawing")
	input := flag.String("input", "", "the file of ready to input, just support shp now")
	output := flag.String("output", "", "the file of to be created, just support sqlite now")
	name := flag.String("name", "", "the dataset name of to be created, default to the output file's title")
	datafile := flag.String("datafile", "", "spatial data file, such as shp/sqlite/udbx")
	indextype := flag.String("indextype", "", "spatial index, such as grid/qtree/rtree/zorder/xzorder")
	txtfile := flag.String("txtfile", "", "the file of saving list of SuperMap Mosaic dataset")
	gmrfile := flag.String("gmrfile", "", "gogis mosaic dataset file")
	path := flag.String("path", "", "gogis mosaic dataset file")
	opengl := flag.Bool("opengl", false, "if enable use opengl")

	flag.Parse() //解析输入的参数

	tr := base.NewTimeRecorder()
	switch subCommand {
	case "help":
		fmt.Println(`open source: https://github.com/pampa0629/gogis; 
					 doc: https://docs.qq.com/doc/DT3RCZlptSk55SWtz`)
	case "version":
		fmt.Println("0.1.7")
	case "server":
		Server(*gmsfile, *cachepath)
	case "desktop":
		Desktop(*mapfile, *datafile, *name, *width, *height, *opengl)
	case "cache":
		Cache(*mapfile, *maptype, *cachepath)
	case "drawmap":
		DrawMap(*mapfile, *picfile, *width, *height, *batch)
	case "convert":
		Convert(*input, *output, *name)
	case "createmap":
		CreateMap(*mapfile, *datafile, *name)
	case "createindex":
		CreateIndex(*datafile, *indextype)
	case "updatesqlite":
		UpdateSqlite(*datafile, *indextype)
	case "inputmosaic":
		InputMosaic(*txtfile, *gmrfile)
	case "buildmosaic":
		BuildMosaic(*path, *gmrfile)
	default:
		other()
	}

	tr.Output(subCommand)
	fmt.Println("DONE!")
}

func Desktop(mapfile, datafile, name string, width, height int, opengl bool) {
	ctype := draw.Default
	if opengl {
		ctype = draw.GLSL
	}
	gmap := mapping.NewMap(ctype)
	gmap.Open(mapfile)
	add2Map(gmap, datafile, name)
	if opengl {
		desktop.ShowGL(gmap, width, height)
	} else {
		desktop.ShowKI(gmap, width, height)
	}
}

func Server(gmsfile, cachepath string) {
	server.StartServer(gmsfile, cachepath)
}
