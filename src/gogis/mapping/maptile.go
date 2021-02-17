package mapping

import (
	"errors"
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"math"
	"strconv"
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
	cancel    bool // 是否被用户取消
}

func NewMapTile(amap *Map, espg EPSG) *MapTile {
	var maptile = new(MapTile)
	maptile.amap = amap
	maptile.epsg = espg
	return maptile
}

func (this *MapTile) prepareBars(mapname string, minLevel, maxLevel int32, fun base.ProgressFunc) []*base.Bar {
	pgb := base.NewProgress("正在生成地图缓存:"+mapname, fun)
	bars := make([]*base.Bar, maxLevel-minLevel+1)

	for i := minLevel; i <= maxLevel; i++ {
		minCol, maxCol, minRow, maxRow := calcColRow(int(i), this.amap.BBox, this.epsg)
		msg := "Level:" + strconv.Itoa(int(i))
		bars[i-minLevel] = pgb.NewBar(msg, int64((maxCol-minCol+1)*(maxRow-minRow+1)))
	}
	return bars
}

// 缓存地图；
// path是最上层的目录，mapname是path下面的子目录；
// fun：进度条函数，外部给出，获取内部进度信息；nil为不关心进度信息
func (this *MapTile) Cache(path, mapname string, maptype draw.MapType, fun base.ProgressFunc) {
	// 第一步，计算缓存的层级范围， 先假设都是经纬度数据
	minLevel, maxLevel := this.calcCacheLevels()
	fmt.Println("min & max Level:", minLevel, maxLevel)

	this.tilestore = new(data.FileTileStore)    // data.FileTileStore LeveldbTileStore
	this.tilestore.Open(path, mapname, maptype) //
	defer this.tilestore.Close()

	// 进度条
	bars := this.prepareBars(mapname, minLevel, maxLevel, fun)

	// 第二步，分层级和范围进行并发生成缓存
	var gm *base.GoMax = new(base.GoMax)
	gm.Init(5000) // 涉及文件操作，并发最大值10000
	for i := minLevel; i <= maxLevel; i++ {
		gm.Add()
		go this.CacheOneLevel(int(i), path, maptype, gm, bars[i-minLevel])
		if this.cancel {
			fmt.Println("Cache cancel")
			return
		}
	}
	gm.Wait()
}

// 缓存指定的层级
func (this *MapTile) CacheOneLevel(level int, path string, maptype draw.MapType, gm *base.GoMax, bar *base.Bar) {
	// func (this *MapTile) CacheOneLevel(level int, path string, maptype draw.MapType, bar *base.Bar) {
	defer gm.Done()

	// 根据范围计算下一级子目录
	// 先计算，当前层级每个瓦片的边长
	minCol, maxCol, minRow, maxRow := calcColRow(level, this.amap.BBox, this.epsg)
	// 并行生成 某个瓦片
	for j := minCol; j <= maxCol; j++ {
		for i := minRow; i <= maxRow; i++ {
			gm.Add()
			// 具体生成瓦片文件
			go this.CacheOneTile(level, j, i, maptype, gm, bar)
			if this.cancel {
				fmt.Println("CacheOneLevel cancel")
				return
			}
		}
	}
}

// 具体生成一个瓦片文件
func (this *MapTile) CacheOneTile(level int, col int, row int, maptype draw.MapType, gm *base.GoMax, bar *base.Bar) {
	data, err := this.CacheOneTile2Bytes(level, col, row, maptype)
	if data != nil && err == nil {
		this.tilestore.Put(level, col, row, data)
	}
	if gm != nil {
		defer gm.Done()
	}
	if bar != nil && bar.Add(1) {
		this.cancel = true
		fmt.Println("bar add cancel")
	}
}

// 缓存一个瓦片，返回指定格式的数据切片
func (this *MapTile) CacheOneTile2Bytes(level int, col int, row int, maptype draw.MapType) ([]byte, error) {
	tmap := this.amap.Copy()
	// 这里关键要把 map要绘制的范围设置对了；即根据 level，row，col来计算bbox
	tmap.BBox = CalcBBox(level, col, row, this.epsg)
	// 不相交就直接返回
	if !tmap.BBox.IsIntersects(this.amap.BBox) {
		return nil, errors.New("bbox is not intersect.")
	}
	// fmt.Println(level, col, row, tmap.BBox)

	tmap.Prepare(256, 256)
	// 图片格式
	if maptype.IsImgType() {
		drawCount := tmap.Draw()
		// 只有真正绘制对象了，才缓存为文件
		if drawCount > 0 {
			// 还得看一下 image 中是否赋值了，彻底防止输出全空的图片
			if tmap.canvas.CheckDrawn() {
				return tmap.Output2Bytes(maptype), nil
			}
		}
	} else if maptype == draw.TypeMvt {
		// 非图片格式
		data, count := tmap.OutputMvt()
		if count > 0 {
			return data, nil
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
func (this *MapTile) calcCacheLevels() (minLevel, maxLevel int32) {
	geoCount := int64(0)       // 矢量暂定看对象个数
	bestRes := math.MaxFloat64 // 栅格看最佳分辨率（越小层级越高）
	for _, layer := range this.amap.Layers {
		switch layer.GetType() {
		case LayerFeature:
			feaLayer := layer.(*FeatureLayer)
			geoCount += feaLayer.feaset.GetCount()
		case LayerRaster:
			rasLayer := layer.(*RasterLayer)
			bestRes = math.Min(bestRes, rasLayer.dt.GetResolution())
		}
	}

	minLevel, maxLevel = base.CalcMinMaxLevels(this.amap.BBox, geoCount)
	maxLevel = base.Int32Max(maxLevel, this.calcResLevel(bestRes))
	return
}

// 根据分辨率计算合适的层级
func (this *MapTile) calcResLevel(res float64) int32 {
	level, lRes := int32(0), 180.0/256.0 // 0 层对应的是 180度/256
	for res < lRes {
		level++
		lRes /= 2.0
	}
	return level
}
