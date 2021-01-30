package data

import (
	"encoding/binary"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"io"
	"os"
	"strconv"
	"sync"
)

// 矢量金字塔
type VectorPyramid struct {
	pyramids           [][]geometry.Geometry
	minLevel, maxLevel int32
	levels             map[int32]int32 // 通过level找到pyramids数组的index
}

func (this *VectorPyramid) Clear() {
	this.pyramids = this.pyramids[:0]
	this.levels = make(map[int32]int32, 0)
}

func saveMaps(maps map[int32]int32, w io.Writer) {
	count := int32(len(maps))
	binary.Write(w, binary.LittleEndian, count)
	keys := make([]int32, count)
	values := make([]int32, count)
	pos := 0
	for k, v := range maps {
		keys[pos] = int32(k)
		values[pos] = int32(v)
		pos++
	}
	binary.Write(w, binary.LittleEndian, keys)
	binary.Write(w, binary.LittleEndian, values)
}

func loadMaps(r io.Reader) map[int32]int32 {
	var count int32
	binary.Read(r, binary.LittleEndian, &count)
	keys := make([]int32, count)
	values := make([]int32, count)
	binary.Read(r, binary.LittleEndian, keys)
	binary.Read(r, binary.LittleEndian, values)
	maps := make(map[int32]int32, count)
	for i := int32(0); i < count; i++ {
		maps[keys[i]] = values[i]
	}
	return maps
}

// 每多少条记录写一个data文件
const PRD_GEO_COUNT = 100000

func (this *VectorPyramid) saveOnePyramid(geos []geometry.Geometry, num int, prdPath string, wg *sync.WaitGroup) {
	defer wg.Done()
	goCount := len(geos)/PRD_GEO_COUNT + 1
	var wg2 *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < int(goCount); i++ {
		wg2.Add(1)
		start := i * PRD_GEO_COUNT
		end := base.IntMin(start+PRD_GEO_COUNT, len(geos))
		go this.saveGeos(geos[start:end], num, i, prdPath, wg2)
	}
	wg2.Wait()
}

func (this *VectorPyramid) saveGeos(geos []geometry.Geometry, n1, n2 int, prdPath string, wg2 *sync.WaitGroup) {
	defer wg2.Done()
	dataName := prdPath + strconv.Itoa(n1) + "-" + strconv.Itoa(n2) + ".data"
	w, _ := os.Create(dataName)
	defer w.Close()
	binary.Write(w, binary.LittleEndian, int32(len(geos)))
	for _, v := range geos {
		if v == nil {
			binary.Write(w, binary.LittleEndian, int32(geometry.TGeoEmpty))
		} else {
			data := v.To(geometry.WKB)
			binary.Write(w, binary.LittleEndian, int32(v.Type()))
			binary.Write(w, binary.LittleEndian, int64(v.GetID()))
			binary.Write(w, binary.LittleEndian, int32(len(data)))
			binary.Write(w, binary.LittleEndian, data)
		}
	}
}

// 保存起来；调用者保证目录已经存在，且目录最后带"/"
func (this *VectorPyramid) Save(prdPath string) {
	// 保存为一个目录，包括：
	// 1) 配置文件: config.prd

	prdName := prdPath + "config.prd"
	prd, _ := os.Create(prdName)
	defer prd.Close()
	binary.Write(prd, binary.LittleEndian, this.minLevel)
	binary.Write(prd, binary.LittleEndian, this.maxLevel)
	saveMaps(this.levels, prd)
	binary.Write(prd, binary.LittleEndian, int32(len(this.pyramids)))
	for _, v := range this.pyramids {
		binary.Write(prd, binary.LittleEndian, int32(len(v)))
	}

	// 2) 分层,分ids存储的数据文件 n1-n2.data
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < len(this.pyramids); i++ {
		wg.Add(1)
		go this.saveOnePyramid(this.pyramids[i], i, prdPath, wg)
	}
	wg.Wait()
}

func (this *VectorPyramid) loadOnePyramid(geoss [][]geometry.Geometry, n1, prdCount, geoCount int, prdPath string, wg *sync.WaitGroup) {
	defer wg.Done()

	geoss[n1] = make([]geometry.Geometry, geoCount)
	goCount := geoCount/PRD_GEO_COUNT + 1
	var wg2 *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < goCount; i++ {
		wg2.Add(1)
		go this.loadGeos(geoss[n1][i*PRD_GEO_COUNT:], n1, i, prdPath, wg2)
	}
	wg2.Wait()
}

func (this *VectorPyramid) loadGeos(geos []geometry.Geometry, n1, n2 int, prdPath string, wg2 *sync.WaitGroup) {
	defer wg2.Done()
	dataName := prdPath + strconv.Itoa(n1) + "-" + strconv.Itoa(n2) + ".data"
	r, _ := os.Open(dataName)
	defer r.Close()

	var geoCount int32
	binary.Read(r, binary.LittleEndian, &geoCount)
	for i := int32(0); i < geoCount; i++ {
		var geotype, length int32
		var id int64
		binary.Read(r, binary.LittleEndian, &geotype)
		binary.Read(r, binary.LittleEndian, &id)
		if geometry.GeoType(geotype) != geometry.TGeoEmpty {
			binary.Read(r, binary.LittleEndian, &length)
			data := make([]byte, length)
			binary.Read(r, binary.LittleEndian, data)
			geo := geometry.CreateGeo(geometry.GeoType(geotype))
			geo.From(data, geometry.WKB)
			geos[i] = geo
		}
	}
}

func (this *VectorPyramid) Load(prdPath string) {
	prdName := prdPath + "config.prd"
	prd, _ := os.Open(prdName)
	defer prd.Close()

	binary.Read(prd, binary.LittleEndian, &this.minLevel)
	binary.Read(prd, binary.LittleEndian, &this.maxLevel)
	this.levels = loadMaps(prd)
	var prdCount int32
	binary.Read(prd, binary.LittleEndian, &prdCount)
	this.pyramids = make([][]geometry.Geometry, prdCount)

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := int32(0); i < prdCount; i++ {
		var geoCount int32
		binary.Read(prd, binary.LittleEndian, &geoCount)

		wg.Add(1)
		// this.pyramids[i] = make([]geometry.Geometry, geoCount)
		go this.loadOnePyramid(this.pyramids, int(i), int(prdCount), int(geoCount), prdPath, wg)
	}
	wg.Wait()
}

// 构建矢量金字塔，应固定比例尺出图的需要进行数据抽稀，根据两个vertex是否在一个像素来决定取舍
func (this *VectorPyramid) Build(bbox base.Rect2D, features []Feature) {
	this.minLevel, this.maxLevel = base.CalcMinMaxLevels(bbox, int64(len(features)))
	fmt.Println("min & max Level:", this.minLevel, this.maxLevel)

	const LEVEL_GAP = 3 // 每隔若干层级构建一个矢量金字塔
	gap := int((this.maxLevel-this.minLevel)/LEVEL_GAP) + 1
	fmt.Println("矢量金字塔层数：", gap)
	this.pyramids = make([][]geometry.Geometry, gap)
	this.levels = make(map[int32]int32)
	index := int32(0)
	for level := this.minLevel; level <= this.maxLevel; level += LEVEL_GAP {
		this.levels[level] = index
		this.pyramids[index] = make([]geometry.Geometry, len(features))
		index++
	}

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for level := this.minLevel; level <= this.maxLevel; level += LEVEL_GAP {
		wg.Add(1)
		// 这里先计算level对应的每个像素的经纬度距离
		dis := base.CalcLevelDis(int(level))
		dis2 := dis * dis
		fmt.Println("层级，距离：", level, dis)
		go this.thinOneLevel(int(level), dis2, features, wg)
	}
	wg.Wait()

	fmt.Println("VectorPyramid.Build()", this.minLevel, this.maxLevel)
}

// 抽稀时，一个批量处理对象个数
const ONE_THIN_COUNT = 100000

// 抽稀一个层级
func (this *VectorPyramid) thinOneLevel(level int, dis2 float64, features []Feature, wg *sync.WaitGroup) {
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
		go this.thinBatch(i*ONE_THIN_COUNT, count, level, features, dis2, wg2)
	}
	wg2.Wait()
}

func (this *VectorPyramid) thinBatch(start, count, level int, features []Feature, dis2 float64, wg2 *sync.WaitGroup) {
	defer wg2.Done()
	index := this.levels[int32(level)]
	end := start + count
	for i := start; i < end; i++ {
		this.pyramids[index][i] = features[i].Geo.Thin(dis2, 120.0)
	}
}
