// 地图服务类

package server

import (
	"bufio"
	"bytes"
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/mapping"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// 所有地图服务的统一管理器
// todo 增加瓦片的内存缓存
type MapServices struct {
	mapServices map[string]*MapService
	cachepath   string
}

func (this *MapServices) Close() {
	for _, v := range this.mapServices {
		v.Close()
	}
}

// 打开地图服务统一配置文件gms；若配置文件不存在，则为数据目录，自动创建gmap对象
// cachepath 为缓存存放路径，若为""或不存在，则默认为 ./cache/
func (this *MapServices) Open(gmsfile string, cachepath string) {
	this.cachepath = cachepath

	if base.FileIsExist(gmsfile) {
		this.open(gmsfile, cachepath)
	} else {
		// 文件不存在，那就取出路径来
		datapath := path.Dir(gmsfile)
		// 扫描路径，默认配置所有已知的数据源
		this.scan(datapath, cachepath)
		// 扫描完了要存起来
		gmsfile = base.GetAbsolutePath(datapath+"/", "gogis."+base.EXT_MAP_SERVICE_FILE)
		gmpfiles := make([]string, 0)
		for _, v := range this.mapServices {
			gmpfiles = append(gmpfiles, v.mapfile)
		}
		writeGmsFile(gmsfile, gmpfiles)
	}
}

// 当已经有gms配置文件时，打开并读取
func (this *MapServices) open(gmsfile string, cachepath string) {
	mapfiles := readGmsFile(gmsfile)
	this.mapServices = make(map[string]*MapService, len(mapfiles))
	for _, mapfile := range mapfiles {
		mapService := new(MapService)
		mapService.OpenGmp(mapfile, cachepath)
		this.mapServices[mapService.maptitle] = mapService
	}
}

// 当没有配置文件时，根据路径自动搜索数据文件，并构造map及其服务
func (this *MapServices) scan(datapath string, cachepath string) {
	this.mapServices = make(map[string]*MapService, 0)
	filepath.Walk(datapath, this.WalkDS)
}

func (this *MapServices) WalkDS(filepath string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if f == nil || f.IsDir() {
		return nil
	}

	// todo 未来应识别更多数据文件
	if path.Ext(filepath) == ".shp" {
		mapService := new(MapService)
		mapService.OpenShp(filepath, this.cachepath)
		this.mapServices[mapService.maptitle] = mapService
	}

	return nil
}

// 读取配置文件，构造出map文件的绝对路径
func readGmsFile(gmsfile string) (mapfiles []string) {
	f, _ := os.Open(gmsfile)
	br := bufio.NewReader(f)
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		strline := string(line)
		strline = strings.TrimSpace(strline)
		// 支持注释掉一整行
		if !(len(strline) >= 2 && strline[0:2] == "//") {
			mapfile := base.GetAbsolutePath(gmsfile, strline)
			fmt.Println("map file:", mapfile)
			mapfiles = append(mapfiles, mapfile)
		}
	}
	return
}

// 写入 gms地图服务配置文件
func writeGmsFile(gmsfile string, gmpfiles []string) {
	f, _ := os.Create(gmsfile)
	for _, gmp := range gmpfiles {
		relPath := base.GetRelativePath(gmsfile, gmp)
		f.WriteString(relPath + "\n")
	}
	f.Close()
}

// 得到瓦片，没有就生成，并缓存起来
func (this *MapServices) GetTile(mapname string, level int, col int, row int, epsg mapping.EPSG) (data []byte) {
	mapService := this.mapServices[mapname]
	if mapService != nil {
		data = mapService.GetTile(level, col, row, epsg)
	}
	return
}

const CACHE_FILE_SIZE = 1024 * 1024 * 100

// 单个地图服务
type MapService struct {
	mapfile   string
	maptitle  string
	gmap      *mapping.Map // 实际干活的gogis map对象
	cachefile string       // 缓存文件
	// db        *leveldb.DB
	tilestore data.TileStore
}

// 打开gms地图配置文件
func (this *MapService) OpenGmp(mapfile string, cachepath string) {
	this.mapfile = mapfile
	this.maptitle = base.GetTitle(mapfile)
	this.gmap = mapping.NewMap()
	this.gmap.Open(mapfile)

	this.OpenCache(cachepath)
}

// 打开 shape文件，生成地图服务
// todo 未来应识别更多数据文件
func (this *MapService) OpenShp(shpfile string, cachepath string) {
	feaset := data.OpenShape(shpfile)
	// 创建地图
	this.gmap = mapping.NewMap()
	this.gmap.AddLayer(feaset)
	// 保存地图文件
	this.mapfile = strings.TrimSuffix(shpfile, ".shp") + "." + base.EXT_MAP_FILE
	this.gmap.Save(this.mapfile)
	this.maptitle = base.GetTitle(this.mapfile)

	this.OpenCache(cachepath)
}

func (this *MapService) OpenCache(cachepath string) {
	this.tilestore = new(data.LeveldbTileStore) //  LeveldbTileStore FileTileStore
	this.tilestore.Open(cachepath, this.maptitle)
}

func (this *MapService) GetTile(level int, col int, row int, epsg mapping.EPSG) (data []byte) {
	bbox := mapping.CalcBBox(level, col, row, epsg)
	fmt.Println("map:", this.maptitle, "level=", level, "col=", col, "row=", row, "BBox:", bbox)

	if !this.gmap.BBox.IsIntersect(bbox) {
		return
	}

	data = this.tilestore.Get(level, col, row)
	// 没有缓存，就必须先生成缓存
	if data == nil || len(data) == 0 {
		// 这里根据地图名字，输出并返回图片
		mapTile := mapping.NewMapTile(this.gmap, epsg)

		tmap, _ := mapTile.CacheOneTile2Map(level, col, row, nil)
		if tmap != nil {
			buf := bytes.NewBuffer(data)
			png.Encode(buf, tmap.OutputImage())
			data = buf.Bytes()
			this.tilestore.Put(level, col, row, data)
		}
		// else {
		// 	// fmt.Println("tile map is nil, error is:", err)
		// }
	}
	return data
}

func (this *MapService) Close() {
	this.tilestore.Close()
	this.gmap.Close()
}
