package mapping

import (
	"fmt"
	"gogis/base"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// 按照固定比例尺来缓存地图瓦片
/* 参考 epsg4326 的方式
目录：Level/col/row.png
Level 0：180度
Level 1：90度
Level 2: 45度
......

*/
func (this *Map) Cache(path string) {
	// 第一步，计算缓存的层级范围， 先假设都是经纬度数据
	minLevel, maxLevel := this.calcCacheLevels()
	fmt.Println("min & max Level:", minLevel, maxLevel)

	// 创建根目录
	os.MkdirAll(path, os.ModePerm)

	// 第二步，分层级和范围进行并发生成缓存
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	// for i := minLevel; i <= maxLevel; i++ {
	for i := minLevel; i <= minLevel+3; i++ {
		wg.Add(1)
		go this.CacheOneLevel(i, path, wg)
	}
	wg.Wait()
}

// 缓存指定的层级
func (this *Map) CacheOneLevel(level int, path string, wg *sync.WaitGroup) {
	defer wg.Done()

	// 1，创建该层级根目录
	levelPath := filepath.Join(path, strconv.Itoa(level))
	os.MkdirAll(levelPath, os.ModePerm)

	// 2，根据范围计算下一级子目录
	// 先计算，当前层级每个瓦片的边长
	dis := 180.0 / (math.Pow(2, float64(level)))

	// epsg 4326（wgs84），瓦片的起点在（0，90）
	// 列的范围，从左到右
	minCol := int(math.Floor((this.BBox.Min.X) / dis))
	maxCol := int(math.Ceil((this.BBox.Max.X) / dis))

	// 行的范围: 瓦片的方向是从上到下
	minRow := int(math.Floor((90.0 - this.BBox.Max.Y) / dis))
	maxRow := int(math.Ceil((90.0 - this.BBox.Min.Y) / dis))

	// 开始生成缓存
	fmt.Printf("开始生成第%v层瓦片，行范围：%v-%v，列范围：%v-%v", level, minRow, maxRow, minCol, maxCol)
	fmt.Println("")

	// 3，并行生成 某个瓦片
	for j := minCol; j <= maxCol; j++ {
		colPath := filepath.Join(levelPath, strconv.Itoa(j))
		os.MkdirAll(colPath, os.ModePerm)
		var wg2 *sync.WaitGroup = new(sync.WaitGroup)
		for i := minRow; i <= maxRow; i++ {
			wg2.Add(1)
			// 具体生成瓦片
			go this.CacheOneTile2File(level, j, i, colPath, wg2)
		}
		wg2.Wait()
		base.DeleteEmptyDir(colPath)
	}

	base.DeleteEmptyDir(levelPath)
}

// 具体生成一个瓦片文件
func (this *Map) CacheOneTile2File(level int, col int, row int, path string, wg *sync.WaitGroup) {
	filename := path + "/" + strconv.Itoa(row) + ".png"

	tmap := this.CacheOneTile2Map(level, col, row, wg)
	if tmap != nil {
		tmap.Output2File(filename, "png")
	}
}

// 缓存一个瓦片，返回Map（无效返回 nil）
func (this *Map) CacheOneTile2Map(level int, col int, row int, wg *sync.WaitGroup) *Map {
	if wg != nil {
		defer wg.Done()
	}

	tmap := this.Copy()
	// 这里关键要把 map要绘制的范围设置对了；即根据 level，row，col来计算bbox
	tmap.BBox = this.calcBBox(level, row, col)
	// 不想交就直接返回
	if !tmap.BBox.IsIntersect(this.BBox) {
		return nil
	}

	tmap.Prepare(256, 256)
	drawCount := tmap.Draw()
	// fmt.Println("after draw, tile image:", this.canvas.img)

	// 只有真正绘制对象了，才缓存为文件
	if drawCount > 0 {
		// 还得看一下 image中是否 赋值了
		if tmap.checkDrawn() {
			return tmap
		}
	}
	return nil
}

// 检查是否真正在image中赋值了
func (this *Map) checkDrawn() bool {
	// return true
	if this.canvas.img.Pix != nil {
		// fmt.Println("tile image:", this.canvas.img)
		for _, v := range this.canvas.img.Pix {
			if v != 0 {
				return true
			}
		}
	}
	return false
}

func (this *Map) calcBBox(level int, row int, col int) (bbox base.Rect2D) {
	dis := 180.0 / (math.Pow(2, float64(level)))
	bbox.Min.X = float64(col) * dis
	bbox.Max.X = bbox.Min.X + dis

	bbox.Max.Y = 90 - float64(row)*dis
	bbox.Min.Y = bbox.Max.Y - dis
	return
}

// 根据bbox和对象数量，计算缓存的最小最大合适层级
// 再小的层级没有必要（图片上的显示范围太小）；再大的层级则瓦片上对象太稀疏
func (this *Map) calcCacheLevels() (minLevel, maxLevel int) {
	geoCount := 0
	for _, layer := range this.Layers {
		geoCount += layer.feaset.Count()
	}

	return base.CalcMinMaxLevels(this.BBox, geoCount)
}
