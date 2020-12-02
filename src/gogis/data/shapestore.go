package data

import (
	"errors"
	"fmt"
	"gogis/base"
	"strconv"
	"strings"
	"sync"
)

// 快捷方法，打开一个shape文件，得到特征集对象
func OpenShape(filename string) Featureset {
	shp := new(ShapeStore)
	params := NewConnParams()
	params["filename"] = filename
	shp.Open(params)
	feaset, _ := shp.GetFeasetByNum(0)
	return feaset
}

// todo 未来还要考虑实现打开一个文件夹
type ShapeStore struct {
	feaset *ShapeFeaset
	name   string //  filename
}

// 打开一个shape文件，params["filename"] = "c:/data/a.shp"
func (this *ShapeStore) Open(params ConnParams) (bool, error) {
	this.feaset = new(ShapeFeaset)
	this.feaset.store = this
	this.name = params["filename"]
	return this.feaset.Open(this.name)
}

func (this *ShapeStore) GetConnParams() ConnParams {
	params := NewConnParams()
	params["filename"] = this.name
	params["type"] = string(this.GetType())
	return params
}

// 得到存储类型
func (this *ShapeStore) GetType() StoreType {
	return StoreShape
}

func (this *ShapeStore) GetFeasetByNum(num int) (Featureset, error) {
	if num == 0 {
		return this.feaset, nil
	}
	return nil, errors.New(strconv.Itoa(num) + " beyond the num of feature sets")
}

func (this *ShapeStore) GetFeasetByName(name string) (Featureset, error) {
	if this.feaset.name == name {
		return this.feaset, nil
	}
	return nil, errors.New("feature set: " + name + " cannot find")
}

func (this *ShapeStore) FeaturesetNames() []string {
	names := make([]string, 1)
	names[0] = this.feaset.name
	return names
}

// 关闭，释放资源
func (this *ShapeStore) Close() {
	// this.MemoryStore.Close()
	this.feaset.Close()
}

// shape数据集，内置内存数据集
type ShapeFeaset struct {
	MemFeaset
	store *ShapeStore
}

// 打开shape文件
func (this *ShapeFeaset) Open(filename string) (bool, error) {
	this.name = base.GetTitle(filename)

	shape := new(ShapeFile)
	res := shape.Open(filename)
	this.bbox = base.NewRect2D(shape.Xmin, shape.Ymin, shape.Xmax, shape.Ymax)

	this.loadShape(shape)
	shape.Close()

	//  处理空间索引文件
	this.loadSpatialIndex()

	//  todo 矢量金字塔
	// this.BuildPyramids()

	return res, nil
}

// 创建或者加载空间索引文件
func (this *ShapeFeaset) loadSpatialIndex() {
	indexName := strings.TrimSuffix(this.store.name, ".shp") + "." + GIX_MARK
	if base.FileIsExist(indexName) {
		this.index = loadGix(indexName)
	} else {
		indexType := TypeGridIndex // 这里确定索引类型 TypeRTreeIndex
		index := this.BuildSpatialIndex(indexType)
		saveGix(indexName, index)
	}
}

func (this *ShapeFeaset) GetStore() Datastore {
	return this.store
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
