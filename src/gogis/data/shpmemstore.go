package data

import (
	"errors"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"gogis/index"
	"os"
	"strconv"
	"strings"
	"sync"
)

// 快捷方法，打开一个shape文件，得到要素集对象
func OpenShape(filename string) Featureset {
	// 默认用内存模式
	shp := new(ShpmemStore)
	params := NewConnParams()
	params["filename"] = filename
	shp.Open(params)
	feaset, _ := shp.GetFeasetByNum(0)
	feaset.Open()
	return feaset
}

func init() {
	RegisterDatastore(StoreShapeMemory, NewShpmemStore)
}

func NewShpmemStore() Datastore {
	return new(ShpmemStore)
}

// todo 未来还要考虑实现打开一个文件夹
// 内存模式的shape存储库
type ShpmemStore struct {
	feaset   *ShpmemFeaset
	filename string
}

// 打开一个shape文件，params["filename"] = "c:/data/a.shp"
func (this *ShpmemStore) Open(params ConnParams) (bool, error) {
	this.filename = params["filename"].(string)
	this.feaset = new(ShpmemFeaset)
	this.feaset.store = this
	this.feaset.filename = this.filename
	this.feaset.name = base.GetTitle(this.filename)
	return true, nil
}

func (this *ShpmemStore) GetConnParams() ConnParams {
	params := NewConnParams()
	params["filename"] = this.filename
	params["type"] = string(this.GetType())
	return params
}

// todo
func (this *ShpmemStore) CreateFeaset(name string, bbox base.Rect2D, geotype geometry.GeoType) Featureset {
	return nil
}

// 得到存储类型
func (this *ShpmemStore) GetType() StoreType {
	return StoreShapeMemory
}

func (this *ShpmemStore) GetFeasetByNum(num int) (Featureset, error) {
	if num == 0 {
		return this.feaset, nil
	}
	return nil, errors.New(strconv.Itoa(num) + " beyond the num of feature sets")
}

func (this *ShpmemStore) GetFeasetByName(name string) (Featureset, error) {
	if strings.ToLower(this.feaset.name) == strings.ToLower(name) {
		return this.feaset, nil
	}
	return nil, errors.New("feature set: " + name + " cannot find")
}

func (this *ShpmemStore) GetFeasetNames() []string {
	names := make([]string, 1)
	names[0] = this.feaset.name
	return names
}

// 关闭，释放资源
func (this *ShpmemStore) Close() {
	this.feaset.Close()
}

// 全内存模式的shape数据集
type ShpmemFeaset struct {
	MemFeaset
	filename string
	store    *ShpmemStore
}

// 打开shape文件
func (this *ShpmemFeaset) Open() (bool, error) {
	tr := base.NewTimeRecorder()

	shape := new(ShapeFile)
	res := shape.Open(this.filename)
	this.bbox = base.NewRect2D(shape.Xmin, shape.Ymin, shape.Xmax, shape.Ymax)
	this.geoType = geometry.ShpType2Geo(shape.GeoType)

	// 处理投影坐标系
	this.proj = base.PrjFromWkt(shape.prj)

	this.loadShape(shape)
	shape.Close()

	tr.Output("open shape file: " + this.filename + ",")

	//  处理空间索引文件
	this.loadSpatialIndex()

	// this.loadPyramids()

	return res, nil
}

func (this *ShpmemFeaset) loadPyramids() {
	prdPath := strings.TrimSuffix(this.filename, ".shp") + ".pyramid/"

	if base.DirIsExist(prdPath) {
		this.pyramid = new(VectorPyramid)
		this.pyramid.Load(prdPath)
	} else {
		// 创建目录
		os.MkdirAll(prdPath, os.ModePerm)
		this.pyramid = new(VectorPyramid)
		this.pyramid.Build(this.bbox, this.features)
		this.pyramid.Save(prdPath)
	}
}

// 创建或者加载空间索引文件
func (this *ShpmemFeaset) loadSpatialIndex() {
	indexName := strings.TrimSuffix(this.filename, ".shp") + "." + base.EXT_SPATIAL_INDEX_FILE
	if base.FileIsExist(indexName) {
		this.index = index.LoadGix(indexName)
	} else {
		// 这里确定索引类型 TypeQTreeIndex TypeRTreeIndex TypeGridIndex TypeZOrderIndex
		indexType := index.TypeZOrderIndex
		spatialIndex := this.BuildSpatialIndex(indexType)
		index.SaveGix(indexName, spatialIndex)
	}
}

func (this *ShpmemFeaset) GetStore() Datastore {
	return this.store
}

// 一次性从文件加载到内存的记录个数
const ONE_LOAD_COUNT = 50000

// 用多文件读取的方式，把geometry都转载到内存中
func (this *ShpmemFeaset) loadShape(shape *ShapeFile) {
	// 设置字段信息
	this.fieldInfos = shape.GetFieldInfos()

	// 计算一下，需要加载多少次
	concount := (int)(shape.recordNum/ONE_LOAD_COUNT) + 1
	fmt.Println("load shape file:"+this.name+", concurrent count is:", concount)

	this.features = make([]Feature, shape.recordNum)
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < concount; i++ {
		count := ONE_LOAD_COUNT
		if i == concount-1 { // 最后一次循环，剩余的对象个数
			count = shape.recordNum - ONE_LOAD_COUNT*(concount-1)
		}
		wg.Add(1)
		go shape.BatchLoad(nil, i*ONE_LOAD_COUNT, count, this.features[i*ONE_LOAD_COUNT:], wg)
	}
	wg.Wait()
}
