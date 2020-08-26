package data

import (
	"errors"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"sync"
)

type MemoryStore struct {
	feasets []*MemFeaset
}

// todo
func (this *MemoryStore) Open(params ConnParams) (bool, error) {
	return true, nil
}

func (this *MemoryStore) GetFeatureset(name string) (Featureset, error) {
	for _, v := range this.feasets {
		if v.name == name {
			return v, nil
		}
	}
	return nil, errors.New("feature set: " + name + " cannot find")
}

func (this *MemoryStore) FeaturesetNames() []string {
	names := make([]string, len(this.feasets))
	for i, v := range this.feasets {
		names[i] = v.name
	}
	return names
}

// 关闭，释放资源
func (this *MemoryStore) Close() {
	for _, feaset := range this.feasets {
		feaset.Close()
	}
	this.feasets = this.feasets[:0]
}

type MemFeaItr struct {
	ids    []int      // id数组
	feaset *MemFeaset // 数据集指针
	pos    int        // 当前位置
}

func (this *MemFeaItr) Next() (Feature, bool) {
	if this.pos < len(this.ids) {
		oldpos := this.pos
		this.pos++
		return this.feaset.features[this.ids[oldpos]], true
	} else {
		return *new(Feature), false
	}
}

// 内存矢量数据集
type MemFeaset struct {
	name     string
	bbox     base.Rect2D
	features []Feature             // 几何对象的数组
	index    *GridIndex            // 空间索引
	pyramids [][]geometry.Geometry // 矢量金字塔
}

func (this *MemFeaset) Open(name string) (bool, error) {
	return true, nil
}

func (this *MemFeaset) GetName() string {
	return this.name
}

func (this *MemFeaset) GetBounds() base.Rect2D {
	return this.bbox
}

// 根据空间范围查询，返回范围内geo的ids
func (this *MemFeaset) Query(bbox base.Rect2D) FeatureIterator {
	var feaitr MemFeaItr
	feaitr.feaset = this
	feaitr.ids = make([]int, 0)
	minRow, maxRow, minCol, maxCol := this.index.GetGridNo(bbox)

	// 最后赋值
	for i := minRow; i <= maxRow; i++ { // 高度（y方向）代表行
		for j := minCol; j <= maxCol; j++ {
			feaitr.ids = append(feaitr.ids, this.index.indexs[i][j]...)
		}
	}

	// 这里应该还要去掉重复id todo ......
	return &feaitr
}

// 清空内存数据
func (this *MemFeaset) Close() {
	this.features = this.features[:0]
	this.pyramids = this.pyramids[:0]
	// 清空索引
	if this.index != nil {
		this.index.Clear()
	}
}

// 构建空间索引
func (this *MemFeaset) BuildSpatialIndex() {
	if this.index == nil {
		this.index = new(GridIndex)
		this.index.Init(this.bbox, len(this.features))
		this.index.BuildByFeas(this.features)
	}
}

// 计算索引重复度，为后续有可能增加多级格网做准备
func (this *MemFeaset) calcRepeatability() float64 {
	indexCount := 0.0
	for i := 0; i < this.index.row; i++ {
		for j := 0; j < this.index.col; j++ {
			indexCount += float64(len(this.index.indexs[i][j]))
		}
	}
	repeat := indexCount / float64(len(this.features))
	fmt.Println("shp index重复度为:", repeat)
	return repeat
}

// 构建矢量金字塔，应固定比例尺出图的需要进行数据抽稀，根据两个vertex是否在一个像素来决定取舍
func (this *MemFeaset) BuildPyramids() {

	minLevel, maxLevel := base.CalcMinMaxLevels(this.bbox, len(this.features))
	// maxLevel += 3
	fmt.Println("min & max Level:", minLevel, maxLevel)

	const LEVEL_GAP = 3 // 每隔若干层级构建一个矢量金字塔
	gap := int((maxLevel - minLevel) / LEVEL_GAP)
	fmt.Println("矢量金字塔层数：", gap)
	this.pyramids = make([][]geometry.Geometry, gap)
	// this.geoCounts = make(map[int]int, gap)
	// fmt.Println("对象个数：", this.recordNum)

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for level := maxLevel - LEVEL_GAP; level >= minLevel; level -= LEVEL_GAP {
		wg.Add(1)
		// 这里先计算level对应的每个像素的经纬度距离
		dis := base.CalcLevelDis(level)
		fmt.Println("层级，距离：", level, dis)
		go this.thinOneLevel(level, dis, wg)
	}
	wg.Wait()
	// fmt.Println("矢量金字塔层数：", len(this.geoPyms))
}

// 抽稀时，一个批量处理对象个数
const ONE_THIN_COUNT = 100000

// 抽稀一个层级
func (this *MemFeaset) thinOneLevel(level int, dis float64, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("开始构建矢量金字塔，层级为：", level)

	// geometrys := make([]geometry.Geometry, len(this.features))
	forcount := (int)(len(this.features)/ONE_THIN_COUNT) + 1

	var wg2 *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < forcount; i++ {
		wg2.Add(1)
		// go this.thinBatch(i*ONE_THIN_COUNT, level, geometrys, dis, wg2)
	}
	wg2.Wait()

	// todo
	// this.geoPyms[level] = geometrys
	// fmt.Println("对象个数为：", len(geometrys))
	// for i := 0; i < len(geometrys); i++ {
	// 	if geometrys[i] != nil {
	// 		this.geoCounts[level]++
	// 	}
	// }

	// fmt.Println("矢量金字塔对象个数：", level, this.geoCounts[level])
}

// func (this *MemFeaset) thinBatch(num int, level int, geometrys []*shpPolyline, dis float64, wg *sync.WaitGroup) {
// 	defer wg.Done()

// 	for i := 0; i < ONE_THIN_COUNT && num < this.recordNum; i++ {
// 		geometrys[num] = thinOneGeo(this.geometrys[num], dis)
// 		num++
// 	}
// }

// // 对一个geometry进行抽稀
// func thinOneGeo(geo *geometry.Geometry, dis float64) *shpPolyline {
// 	// fmt.Println("thinOneGeo：", id)

// 	var newGeo shpPolyline
// 	newGeo.shpType = geo.shpType
// 	newGeo.box = geo.box
// 	newGeo.numParts = geo.numParts
// 	newGeo.parts = make([]int32, 0, geo.numParts)
// 	newGeo.points = make([]shpPoint, 0)
// 	pos := int32(0)

// 	for i := int32(0); i < geo.numParts; i++ {
// 		var points []shpPoint
// 		if i != geo.numParts-1 {
// 			points = geo.points[geo.parts[i]:geo.parts[i+1]]
// 		} else {
// 			points = geo.points[geo.parts[i]:]
// 		}
// 		newPnts := thinOnePart(points, dis)
// 		newGeo.points = append(newGeo.points, newPnts...)
// 		newGeo.parts = append(newGeo.parts, pos)
// 		pos += int32(len(newPnts))
// 	}
// 	newGeo.numPoints = int32(len(newGeo.points))

// 	if len(newGeo.points) == 0 {
// 		newGeo.numParts = 1
// 		newGeo.numPoints = 2
// 		newGeo.parts = make([]int32, 1)
// 		newGeo.parts[0] = 2
// 		newGeo.points = make([]shpPoint, 2)
// 		copy(newGeo.points, geo.points[0:2])
// 	}
// 	return &newGeo
// }

// func thinOnePart(points []shpPoint, dis float64) (newPnts []shpPoint) {
// 	newPnts = make([]shpPoint, 1, len(points))
// 	dis2 := math.Pow(dis, 2)
// 	newPnts[0] = points[0]
// 	pos := 0

// 	for i := 1; i < len(points); i++ {
// 		if base.DistanceSquare(points[pos].X, points[pos].Y, points[i].X, points[i].Y) > dis2 {
// 			newPnts = append(newPnts, points[i])
// 			pos = i
// 		}
// 	}
// 	// 如果就一个点，清空之
// 	if len(newPnts) == 1 {
// 		newPnts = newPnts[:0]
// 	}
// 	return newPnts
// }
