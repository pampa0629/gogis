package data

import (
	"fmt"
	"gogis/base"
	"sync"
)

type ShapeStore struct {
	MemoryStore // 匿名组合，实现继承效果
	feaset      ShapeFeaset
}

// 打开一个shape文件，params["filename"] = "c:/temp/a.shp"
func (this *ShapeStore) Open(params ConnParams) (bool, error) {
	this.feasets = make([]*MemFeaset, 1)
	this.feasets[0] = &this.feaset.MemFeaset
	return this.feaset.Open(params["filename"])
}

// 关闭，释放资源
func (this *ShapeStore) Close() {
	this.MemoryStore.Close()
	this.feaset.Close()
}

// shape数据集，内置内存数据集
type ShapeFeaset struct {
	MemFeaset
}

// 打开shape文件
func (this *ShapeFeaset) Open(filename string) (bool, error) {
	fmt.Println("ShapeFeaset.Open()")
	shape := new(ShapeFile)
	res := shape.Open(filename)
	this.bbox = base.NewRect2D(shape.Xmin, shape.Ymin, shape.Xmax, shape.Ymax)

	this.loadShape(shape)
	shape.Close()

	this.BuildSpatialIndex()
	// this.BuildPyramids()

	return res, nil
}

// 一次性从文件加载到内存的记录个数
const ONE_LOAD_COUNT = 50000

// 用多文件读取的方式，把geometry都转载到内存中
func (this *ShapeFeaset) loadShape(shape *ShapeFile) {
	// 设置字段信息
	this.fieldInfos = shape.GetFieldInfos()

	// 计算一下，需要加载多少次
	forcount := (int)(shape.recordNum/ONE_LOAD_COUNT) + 1
	fmt.Println("ShapeFeaset.load(), for count: ", forcount)

	this.features = make([]Feature, shape.recordNum)
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < forcount; i++ {
		count := ONE_LOAD_COUNT
		if i == forcount-1 { // 最后一次循环，剩余的对象个数
			count = shape.recordNum - ONE_LOAD_COUNT*(forcount-1)
		}
		wg.Add(1)
		go shape.BatchLoad(i*ONE_LOAD_COUNT, count, this.features[i*ONE_LOAD_COUNT:], wg)
	}
	wg.Wait()
}
