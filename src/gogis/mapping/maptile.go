package mapping

import (
	"bytes"
	"errors"
	"fmt"
	"gogis/base"
	"gogis/data"
	"image/png"
	"math"
	"sync"
)

// 定义坐标系常量
type EPSG int

const (
	Epsg3857 EPSG = 3857 //Web墨卡托
	Epsg4326 EPSG = 4326 // 经纬度
)

// epsg4326：原点在(-180,90)，x轴往右，y轴往下
/* 文件格式：Level/col/row.png
Level 0：180度
Level 1：90度
Level 2: 45度
......
*/

// 地图瓦片（栅格）生成器
type MapTile struct {
	amap      *Map
	epsg      EPSG
	tilestore data.TileStore
}

func NewMapTile(amap *Map, espg EPSG) *MapTile {
	var maptile = new(MapTile)
	maptile.amap = amap
	maptile.epsg = espg
	return maptile
}

// 缓存这个地图
// path是最上层的目录，mapname是path下面的子目录
func (this *MapTile) Cache(path string, mapname string) {
	// 第一步，计算缓存的层级范围， 先假设都是经纬度数据
	minLevel, maxLevel := this.calcCacheLevels()
	fmt.Println("min & max Level:", minLevel, maxLevel)

	this.tilestore = new(data.LeveldbTileStore) // data.FileTileStore LeveldbTileStore
	this.tilestore.Open(path, mapname)
	// 创建根目录
	// os.MkdirAll(path, os.ModePerm)

	// 第二步，分层级和范围进行并发生成缓存
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := minLevel; i <= maxLevel; i++ {
		// for i := minLevel; i <= minLevel+3; i++ {
		wg.Add(1)
		go this.CacheOneLevel(i, path, wg)
	}
	wg.Wait()

	this.tilestore.Close()
}

// 缓存指定的层级
func (this *MapTile) CacheOneLevel(level int, path string, wg *sync.WaitGroup) {
	defer wg.Done()

	// 1，创建该层级根目录
	// levelPath := filepath.Join(path, strconv.Itoa(level))
	// os.MkdirAll(levelPath, os.ModePerm)

	// 2，根据范围计算下一级子目录
	// 先计算，当前层级每个瓦片的边长
	minCol, maxCol, minRow, maxRow := calcColRow(level, this.amap.BBox, this.epsg)

	// 开始生成缓存
	fmt.Printf("开始生成第%v层瓦片，行范围：%v-%v，列范围：%v-%v", level, minRow, maxRow, minCol, maxCol)
	fmt.Println("")

	// 3，并行生成 某个瓦片
	for j := minCol; j <= maxCol; j++ {
		// colPath := filepath.Join(levelPath, strconv.Itoa(j))
		// os.MkdirAll(colPath, os.ModePerm)
		var wg2 *sync.WaitGroup = new(sync.WaitGroup)
		for i := minRow; i <= maxRow; i++ {
			wg2.Add(1)
			// 具体生成瓦片文件
			go this.CacheOneTile2File(level, j, i, wg2)
		}
		wg2.Wait()
		// base.DeleteEmptyDir(colPath)
	}

	// base.DeleteEmptyDir(levelPath)
}

// 具体生成一个瓦片文件
func (this *MapTile) CacheOneTile2File(level int, col int, row int, wg *sync.WaitGroup) {
	// defer wg.Done()
	// filename := path + "/" + strconv.Itoa(row) + ".png"

	tmap, _ := this.CacheOneTile2Map(level, col, row, wg)
	if tmap != nil {
		data := make([]byte, 0)
		buf := bytes.NewBuffer(data)
		png.Encode(buf, tmap.OutputImage())
		data = buf.Bytes()
		this.tilestore.Put(level, col, row, data)
	}

	// if tmap != nil {
	// 	tmap.canvas.Image()
	// 	tmap.Output2File(filename, "png")
	// }
}

// 缓存一个瓦片，返回Map（无效返回 nil）
func (this *MapTile) CacheOneTile2Map(level int, col int, row int, wg *sync.WaitGroup) (*Map, error) {
	if wg != nil {
		defer wg.Done()
	}

	tmap := this.amap.Copy()
	// 这里关键要把 map要绘制的范围设置对了；即根据 level，row，col来计算bbox
	tmap.BBox = CalcBBox(level, col, row, this.epsg)
	// 不相交就直接返回
	if !tmap.BBox.IsIntersect(this.amap.BBox) {
		return nil, errors.New("bbox is not intersect.")
	}

	tmap.Prepare(256, 256)
	drawCount := tmap.Draw()
	// fmt.Println("after draw, tile image:", this.canvas.img)

	// 只有真正绘制对象了，才缓存为文件
	if drawCount > 0 {
		// 还得看一下 image中是否 赋值了
		if tmap.canvas.CheckDrawn() {
			return tmap, nil
		}
	}
	return nil, errors.New("draw count is zeor.")
}

// 根据层级和边框范围，计算得到最大、最小行列数
func calcColRow(level int, bbox base.Rect2D, espg EPSG) (minCol, maxCol, minRow, maxRow int) {
	if espg == Epsg4326 {
		// 先计算，当前层级每个瓦片的边长
		dis := 180.0 / (math.Pow(2, float64(level)))

		// epsg 4326（wgs84），瓦片的起点在（-180，90）
		// 列的范围，从左到右
		minCol = int(math.Floor((bbox.Min.X + 180.0) / dis))
		maxCol = int(math.Ceil((bbox.Max.X + 180.0) / dis))

		// 行的范围: 瓦片的方向是从上到下
		minRow = int(math.Floor((90.0 - bbox.Max.Y) / dis))
		maxRow = int(math.Ceil((90.0 - bbox.Min.Y) / dis))
	}
	return
}

// 根据层级与行列号，计算得到边框
func CalcBBox(level int, col int, row int, espg EPSG) (bbox base.Rect2D) {
	if espg == Epsg4326 {
		dis := 180.0 / (math.Pow(2, float64(level)))
		bbox.Min.X = float64(col)*dis - 180.0
		bbox.Max.X = bbox.Min.X + dis

		bbox.Max.Y = 90 - float64(row)*dis
		bbox.Min.Y = bbox.Max.Y - dis
	}
	return
}

// 根据bbox和对象数量，计算缓存的最小最大合适层级
// 再小的层级没有必要（图片上的显示范围太小）；再大的层级则瓦片上对象太稀疏
func (this *MapTile) calcCacheLevels() (minLevel, maxLevel int) {
	geoCount := int64(0)
	for _, layer := range this.amap.Layers {
		geoCount += layer.feaset.GetCount()
	}

	return base.CalcMinMaxLevels(this.amap.BBox, geoCount)
}
