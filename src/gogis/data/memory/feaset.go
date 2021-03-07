package memory

import (
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/index"
	"sync"
)

// 内存矢量数据集
type MemFeaset struct {
	data.FeasetInfo
	Features []data.Feature // 几何对象的数组
	// id2feaPos  map[int64]int      // 从id到features中位置的对应关系
	Index   index.SpatialIndex  // 空间索引
	Pyramid *data.VectorPyramid // 矢量金字塔
	// data.ProjCommon
	store *MemoryStore
	lock  sync.Mutex
}

// 使用前必须调用，设置features和ids的对照关系，便于后续直接通过id访问
// func (this *MemFeaset) Prepare() {
// 	this.id2feaPos = make(map[int64]int, len(this.features))
// 	for i, v := range this.features {
// 		this.id2feaPos[v.Geo.GetID()] = i
// 	}
// }

func (this *MemFeaset) Open() (bool, error) {
	return true, nil
}

func (this *MemFeaset) GetStore() data.Datastore {
	return this.store
}

// 对象个数
func (this *MemFeaset) GetCount() int64 {
	return int64(len(this.Features))
}

// 综合查询
func (this *MemFeaset) Query(def *data.QueryDef) data.FeatureIterator {
	if def == nil {
		def = new(data.QueryDef)
		def.SpatialObj = this.Bbox
	}
	var feaitr MemFeaItr
	feaitr.feaset = this
	feaitr.fields = def.Fields
	// 先根据空间查询条件做筛选
	feaitr.squery.Init(def.SpatialObj, def.SpatialMode)
	ids := feaitr.squery.QueryIds(this.Index)

	// 再解析 where 语句
	var comps data.FieldComps
	comps.Parse(def.Where, this.FieldInfos)
	for _, id := range ids {
		if comps.Match(this.Features[id].Atts) {
			feaitr.ids = append(feaitr.ids, int64(id))
		}
	}

	return &feaitr
}

// 根据空间范围查询，返回范围内geo的ids
// func (this *MemFeaset) QueryByBounds(bbox base.Rect2D) data.FeatureIterator {
// 	var def data.QueryDef
// 	def.SpatialMode = base.Intersects
// 	def.SpatialObj = bbox
// 	temp := this.QueryByDef(def)

// 	feaitr := temp.(*MemFeaItr)
// 	// 看是否需要用金字塔
// 	if this.Pyramid != nil {
// 		level, _ := base.CalcMinMaxLevels(bbox, 0)
// 		minLevel := int32(100)
// 		for k, v := range this.Pyramid.Levels {
// 			// 需要取所有比需要的level大的金字塔中，最小的哪个
// 			if k >= level && k <= minLevel {
// 				feaitr.geoPyramid = &this.Pyramid.Pyramids[v]
// 				minLevel = k
// 			}
// 		}
// 	}

// 	return feaitr
// }

// 清空内存数据
func (this *MemFeaset) Close() {
	this.Features = this.Features[:0]

	// 清空索引
	if this.Index != nil {
		this.Index.Clear()
	}

	// 清空金字塔
	if this.Pyramid != nil {
		this.Pyramid.Clear()
	}
}

func (this *MemFeaset) BeforeWrite(count int64) {
	this.Features = make([]data.Feature, 0, count)
}

// 批量写入数据
func (this *MemFeaset) BatchWrite(feas []data.Feature) {
	this.lock.Lock()
	this.Features = append(this.Features, feas...)
	this.lock.Unlock()
}

// 计算bbox，创建空间索引
func (this *MemFeaset) EndWrite() {
	for _, v := range this.Features {
		this.Bbox = this.Bbox.Union(v.Geo.GetBounds())
	}
	this.BuildSpatialIndex(index.TypeZOrderIndex)
}

// 构建空间索引
func (this *MemFeaset) BuildSpatialIndex(indexType index.SpatialIndexType) index.SpatialIndex {
	if this.Index == nil {
		tr := base.NewTimeRecorder()

		this.Index = index.NewSpatialIndex(indexType)
		this.Index.Init(this.Bbox, int64(len(this.Features)))
		for i, v := range this.Features {
			v.Geo.SetID(int64(i)) // 内存数据，只能用pos当id了
			this.Index.AddGeo(v.Geo)
		}
		check := this.Index.Check()
		fmt.Println("check building spatial index, result is:", check)

		tr.Output("build spatial index")
		return this.Index
	}
	return nil
}
