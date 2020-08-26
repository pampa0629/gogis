package data

import (
	"errors"
	"fmt"
	"gogis/base"
	"sync"
)

type ShapeStore struct {
	feasets []*ShapeFeaset
}

// 打开一个shape文件，params["filename"] = "c:/temp/a.shp"
func (this *ShapeStore) Open(params ConnParams) (bool, error) {
	// this.MemoryStore.Open(params)
	this.feasets = make([]*ShapeFeaset, 1)
	this.feasets[0] = new(ShapeFeaset)
	return this.feasets[0].Open(params["filename"])
}

func (this *ShapeStore) GetFeatureset(name string) (Featureset, error) {
	for _, v := range this.feasets {
		if v.name == name {
			return v, nil
		}
	}
	return nil, errors.New("feature set: " + name + " cannot find")
}

func (this *ShapeStore) FeaturesetNames() []string {
	names := make([]string, len(this.feasets))
	for i, v := range this.feasets {
		names[i] = v.name
	}
	return names
}

// 关闭，释放资源
func (this *ShapeStore) Close() {
	if this.feasets != nil {
		for _, feaset := range this.feasets {
			feaset.Close()
		}
	}
}

// shape数据集，内置内存数据集
type ShapeFeaset struct {
	MemFeaset
	shapefile ShapeFile
}

// 打开shape文件
func (this *ShapeFeaset) Open(filename string) (bool, error) {
	fmt.Println("ShapeFeaset.Open()")
	res := this.shapefile.Open(filename)
	this.load()
	return res, nil
}

func (this *ShapeFeaset) GetName() string {
	return this.name
}

func (this *ShapeFeaset) GetBounds() base.Rect2D {
	return this.bbox
}

func (this *ShapeFeaset) Query(bbox base.Rect2D) FeatureIterator {
	return this.Query(bbox)
}

// 清空内存数据
func (this *ShapeFeaset) Close() {
	this.Close()
	this.shapefile.Close()
}

// 一次性从文件加载到内存的记录个数
const ONE_LOAD_COUNT = 50000

// 用多文件读取的方式，把geometry都转载到内存中
func (this *ShapeFeaset) load() {
	fmt.Println("ShapeFeaset.load()")
	// 计算一下，需要加载多少次
	forcount := (int)(this.shapefile.recordNum/ONE_LOAD_COUNT) + 1
	fmt.Println("forcount: ", forcount)
	// // fmt.Println("forcount:", forcount)
	// // datas := make([][]byte, forcount)

	// reads := make([]ShpRead, forcount)

	// for i := 0; i < forcount; i++ {
	// 	num := i * ONE_LOAD_COUNT
	// 	// N条记录打包一次性读取
	// 	len := 0
	// 	for i := 0; i < ONE_LOAD_COUNT && num < recordCount; i++ {
	// 		len += (int)(this.shapefile.records[num].length*2 + 8)
	// 		num++
	// 	}
	// 	if i == 0 {
	// 		reads[0].pos = 100
	// 	} else {
	// 		reads[i].pos = reads[i-1].pos + reads[i-1].len
	// 	}

	// 	reads[i].len = len
	// }

	// fmt.Println("reads:", reads)

	this.features = make([]Feature, this.shapefile.recordNum)
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < forcount; i++ {
		count := ONE_LOAD_COUNT
		if i == forcount-1 { // 最后一次循环，剩余的对象个数
			count = this.shapefile.recordNum - ONE_LOAD_COUNT*(forcount-1)
		}
		wg.Add(1)
		go this.shapefile.BatchLoad(i*ONE_LOAD_COUNT, count, this.features[i*ONE_LOAD_COUNT:], wg)
		// go this.loadBatch(i*ONE_LOAD_COUNT, reads[i], wg)
	}

	wg.Wait()
	// this.BuildSpatialIndex()
}
