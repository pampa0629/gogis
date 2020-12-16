package data

import (
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"math"
	"sync"
)

// 矢量金字塔
// todo 提供本地缓存,避免每次都重新生成
type VectorPyramid struct {
	pyramids           [][]geometry.Geometry
	minLevel, maxLevel int
	levels             map[int]int // 通过level 找到 pyramids数组的index
}

func (this *VectorPyramid) Clear() {
	this.pyramids = this.pyramids[:0]
}

// todo 矢量金字塔的构建应该能脱离 具体的引擎类型
// 构建矢量金字塔，应固定比例尺出图的需要进行数据抽稀，根据两个vertex是否在一个像素来决定取舍
func (this *VectorPyramid) Build(bbox base.Rect2D, features []Feature) {
	this.minLevel, this.maxLevel = base.CalcMinMaxLevels(bbox, int64(len(features)))
	fmt.Println("min & max Level:", this.minLevel, this.maxLevel)

	const LEVEL_GAP = 3 // 每隔若干层级构建一个矢量金字塔
	gap := int((this.maxLevel - this.minLevel) / LEVEL_GAP)
	fmt.Println("矢量金字塔层数：", gap)
	this.pyramids = make([][]geometry.Geometry, gap)
	this.levels = make(map[int]int)
	index := 0
	for level := this.maxLevel - LEVEL_GAP; level >= this.minLevel; level -= LEVEL_GAP {
		this.levels[level] = index
		this.pyramids[index] = make([]geometry.Geometry, len(features))
		index++
	}

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for level := this.maxLevel - LEVEL_GAP; level >= this.minLevel; level -= LEVEL_GAP {
		wg.Add(1)
		// 这里先计算level对应的每个像素的经纬度距离
		dis := base.CalcLevelDis(level)
		fmt.Println("层级，距离：", level, dis)
		go this.thinOneLevel(level, dis, features, wg)
	}
	wg.Wait()

	fmt.Println("VectorPyramid.Build()", this.minLevel, this.maxLevel)
}

// 抽稀时，一个批量处理对象个数
const ONE_THIN_COUNT = 100000

// 抽稀一个层级
func (this *VectorPyramid) thinOneLevel(level int, dis float64, features []Feature, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("开始构建矢量金字塔，层级为：", level)
	forcount := (int)(len(features)/ONE_THIN_COUNT) + 1

	var wg2 *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < forcount; i++ {
		count := ONE_THIN_COUNT
		if i == forcount-1 {
			count = len(features) - (forcount-1)*ONE_THIN_COUNT
		}
		wg2.Add(1)
		go this.thinBatch(i*ONE_THIN_COUNT, count, level, features, dis, wg2)
	}
	wg2.Wait()
}

func (this *VectorPyramid) thinBatch(start, count, level int, features []Feature, dis float64, wg2 *sync.WaitGroup) {
	defer wg2.Done()
	index := this.levels[level]
	end := start + count
	for i := start; i < end; i++ {
		this.pyramids[index][i] = thinOneGeo(features[i].Geo, dis)
	}
}

// 对一个geometry进行抽稀
func thinOneGeo(geo geometry.Geometry, dis float64) geometry.Geometry {
	// fmt.Println("thinOneGeo：", id)
	var newgeo geometry.Geometry
	switch geo.Type() {
	case geometry.TGeoPoint:
		// todo .....
	case geometry.TGeoPolyline:
		geo = thinOnePolyline(geo, dis)
	case geometry.TGeoPolygon:
		geo = thinOnePolygon(geo, dis)
	}
	return newgeo
}

func thinOnePolyline(geo geometry.Geometry, dis float64) geometry.Geometry {
	polyline := geo.(*geometry.GeoPolyline)

	var newgeo geometry.GeoPolyline
	newgeo.BBox = polyline.BBox
	newgeo.Points = make([][]base.Point2D, len(polyline.Points))
	for i, v := range polyline.Points {
		newgeo.Points[i] = thinOnePart(v, dis, 2)
	}

	return &newgeo
}

func thinOnePolygon(geo geometry.Geometry, dis float64) geometry.Geometry {
	polygon := geo.(*geometry.GeoPolygon)

	var newgeo geometry.GeoPolygon
	newgeo.BBox = polygon.BBox
	newgeo.Points = make([][][]base.Point2D, len(polygon.Points))
	for i, _ := range polygon.Points {
		newgeo.Points[i] = make([][]base.Point2D, len(polygon.Points[i]))
		for j, v := range polygon.Points[i] {
			newgeo.Points[i][j] = thinOnePart(v, dis, 3)
		}
	}

	return &newgeo
}

func thinOnePart(points []base.Point2D, dis float64, lest int) (newPnts []base.Point2D) {
	newPnts = make([]base.Point2D, 1, len(points))
	dis2 := math.Pow(dis, 2)
	newPnts[0] = points[0]
	pos := 0

	for i := 1; i < len(points); i++ {
		if base.DistanceSquare(points[pos].X, points[pos].Y, points[i].X, points[i].Y) > dis2 {
			newPnts = append(newPnts, points[i])
			pos = i
		}
	}
	// 点数不够，就重复一下
	for len(newPnts) < lest {
		newPnts = append(newPnts, newPnts[0])
	}
	return newPnts
}
