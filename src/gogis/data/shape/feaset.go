// shape存储库，硬盘模式
package shape

import (
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/geometry"
	"gogis/index"
	"strings"
	"sync"
)

// 硬盘读写模式的shape数据集
type ShapeFeaset struct {
	filename string // 文件全路径
	data.FeasetInfo
	store *ShapeStore

	shape *ShapeFile
	// 空间索引
	index index.SpatialIndex

	isCache bool
	geos    []geometry.Geometry // 内存
}

// 打开shape文件
func (this *ShapeFeaset) Open() (bool, error) {
	tr := base.NewTimeRecorder()

	this.shape = new(ShapeFile)
	res := this.shape.Open(this.filename, this.isLoadFields())
	this.Bbox = base.NewRect2D(this.shape.Xmin, this.shape.Ymin, this.shape.Xmax, this.shape.Ymax)
	this.GeoType = geometry.ShpType2Geo(this.shape.GeoType)
	this.FieldInfos = this.shape.GetFieldInfos()
	// 处理投影坐标系
	this.Proj = base.PrjFromWkt(this.shape.prj)

	tr.Output("open shape file: " + this.filename)

	//  处理空间索引文件
	this.loadSpatialIndex()
	tr.Output("load spatial index")

	// 是否缓存到内存中
	cache := this.store.GetConnParams()["cache"]
	if cache != nil && cache.(bool) {
		this.cache()
		tr.Output("cache into memory")
	}

	//  todo 矢量金字塔
	// this.BuildPyramids()
	return res, nil
}

func (this *ShapeFeaset) isLoadFields() bool {
	fields := this.store.GetConnParams()["fields"]
	if fields != nil {
		if fs, ok := fields.([]string); ok {
			if len(fs) == 0 {
				return false
			}
		}
	}
	return true
}

const ONE_LOAD_COUNT = 100000

func (this *ShapeFeaset) cache() {
	this.isCache = true
	this.geos = make([]geometry.Geometry, this.shape.recordNum)
	goCount := this.shape.recordNum/ONE_LOAD_COUNT + 1
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < int(goCount); i++ {
		count := ONE_LOAD_COUNT
		if i == goCount-1 { // 最后一次循环，剩余的对象个数
			count = this.shape.recordNum - ONE_LOAD_COUNT*(goCount-1)
		}
		wg.Add(1)
		go this.gocache(this.geos[i*ONE_LOAD_COUNT:], i*ONE_LOAD_COUNT, count, wg)
	}
	wg.Wait()
}

func (this *ShapeFeaset) gocache(geos []geometry.Geometry, start, count int, wg *sync.WaitGroup) {
	defer wg.Done()
	this.shape.BatchLoad(nil, start, count, geos, nil)
}

func (this *ShapeFeaset) Close() {
	this.shape.Close()
	this.store = nil
}

func (this *ShapeFeaset) GetGeoType() geometry.GeoType {
	// return geometry.ShpType2Geo(this.shape.GeoType)
	return this.GeoType
}

func (this *ShapeFeaset) GetName() string {
	return this.Name
}

// 创建或者加载空间索引文件
func (this *ShapeFeaset) loadSpatialIndex() {
	indexName := strings.TrimSuffix(this.filename, ".shp") + "." + base.EXT_SPATIAL_INDEX_FILE
	if base.FileIsExist(indexName) {
		this.index = index.LoadGix(indexName)
	} else {
		indexType := index.SpatialIndexType(this.store.params.GetString("index"))
		if len(indexType) == 0 {
			indexType = index.TypeQTreeIndex // 这里确定索引类型 TypeQTreeIndex TypeRTreeIndex TypeGridIndex
		}

		spatialIndex := this.BuildSpatialIndex(indexType)
		index.SaveGix(indexName, spatialIndex)
	}
}

func (this *ShapeFeaset) BuildSpatialIndex(indexType index.SpatialIndexType) index.SpatialIndex {
	if this.index == nil {
		tr := base.NewTimeRecorder()

		this.index = index.NewSpatialIndex(indexType)
		this.index.Init(this.Bbox, this.GetCount())
		bboxes, ids := this.shape.LoadBboxIds()
		for i, v := range bboxes {
			this.index.AddOne(v, int64(ids[i]))
		}
		check := this.index.Check()
		fmt.Println("check building spatial index, result is:", check)

		tr.Output("build spatial index")
		return this.index
	}
	return nil
}

func (this *ShapeFeaset) GetStore() data.Datastore {
	return this.store
}

func (this *ShapeFeaset) GetCount() int64 {
	return int64(this.shape.recordNum)
}

func (this *ShapeFeaset) GetBounds() base.Rect2D {
	return this.Bbox
}

func (this *ShapeFeaset) GetFieldInfos() []data.FieldInfo {
	return this.shape.GetFieldInfos()
}

func (this *ShapeFeaset) BeforeWrite(count int64) {

}

// 批量写入数据 todo
func (this *ShapeFeaset) BatchWrite(feas []data.Feature) {
}

func (this *ShapeFeaset) EndWrite() {
}

// func (this *ShapeFeaset) queryByBounds(bbox base.Rect2D) data.FeatureIterator {
// 	feaItr := new(ShapeFeaItr)
// 	feaItr.feaset = this
// 	feaItr.ids = this.index.Query(bbox)
// 	// 给ids排序，以便后面的连续读取
// 	sort.Sort(base.Int64s(feaItr.ids))
// 	return feaItr
// }

func (this *ShapeFeaset) Query(def *data.QueryDef) data.FeatureIterator {
	if def == nil {
		def = new(data.QueryDef)
		def.SpatialObj = this.Bbox
	}
	feaitr := new(ShapeFeaItr)
	feaitr.feaset = this
	feaitr.fields = def.Fields
	feaitr.where = def.Where
	// 先根据空间查询条件做筛选
	feaitr.squery.Init(def.SpatialObj, def.SpatialMode)
	feaitr.ids = feaitr.squery.QueryIds(this.index)
	return feaitr
}
