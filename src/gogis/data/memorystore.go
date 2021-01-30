package data

import (
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"gogis/index"
)

func init() {
	RegisterDatastore(StoreMemory, NewMemoryStore)
}

func NewMemoryStore() Datastore {
	return new(MemoryStore)
}

type MemoryStore struct {
	Feasets
}

// nothing to do
func (this *MemoryStore) Open(params ConnParams) (bool, error) {
	return true, nil
}

func (this *MemoryStore) GetConnParams() ConnParams {
	return NewConnParams()
}

// 得到存储类型
func (this *MemoryStore) GetType() StoreType {
	return StoreMemory
}

// 关闭，释放资源
func (this *MemoryStore) Close() {
	for _, feaset := range this.feasets {
		feaset.Close()
	}
	this.feasets = this.feasets[:0]
}

// todo
func (this *MemoryStore) CreateFeaset(name string, bbox base.Rect2D, geotype geometry.GeoType) Featureset {
	return nil
}

// 内存矢量数据集
type MemFeaset struct {
	name       string
	bbox       base.Rect2D
	geoType    geometry.GeoType
	fieldInfos []FieldInfo
	features   []Feature // 几何对象的数组
	// id2feaPos  map[int64]int      // 从id到features中位置的对应关系
	index   index.SpatialIndex // 空间索引
	pyramid *VectorPyramid     // 矢量金字塔
	projCommon

	// store      *MemoryStore
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

// todo thinking
func (this *MemFeaset) GetStore() Datastore {
	return nil
}

func (this *MemFeaset) GetGeoType() geometry.GeoType {
	return this.geoType
}

// 对象个数
func (this *MemFeaset) GetCount() int64 {
	return int64(len(this.features))
}

func (this *MemFeaset) GetName() string {
	return this.name
}

func (this *MemFeaset) GetBounds() base.Rect2D {
	return this.bbox
}

func (this *MemFeaset) GetFieldInfos() []FieldInfo {
	return this.fieldInfos
}

// 综合查询
func (this *MemFeaset) QueryByDef(def QueryDef) FeatureIterator {
	var feaitr MemFeaItr
	feaitr.feaset = this
	feaitr.fields = def.Fields
	// 先根据空间查询条件做筛选
	feaitr.squery.Init(def.SpatialObj, def.SpatialMode)
	ids := feaitr.squery.QueryIds(this.index)

	// 再解析 where 语句
	var comps FieldComps
	comps.Parse(def.Where, this.fieldInfos)
	for _, id := range ids {
		if comps.Match(this.features[id]) {
			feaitr.ids = append(feaitr.ids, int64(id))
		}
	}

	return &feaitr
}

// 根据空间范围查询，返回范围内geo的ids
func (this *MemFeaset) QueryByBounds(bbox base.Rect2D) FeatureIterator {
	var def QueryDef
	def.SpatialMode = base.Intersects
	def.SpatialObj = bbox
	temp := this.QueryByDef(def)

	feaitr := temp.(*MemFeaItr)
	// 看是否需要用金字塔
	if this.pyramid != nil {
		level, _ := base.CalcMinMaxLevels(bbox, 0)
		minLevel := int32(100)
		for k, v := range this.pyramid.levels {
			// 需要取所有比需要的level大的金字塔中，最小的哪个
			if k >= level && k <= minLevel {
				feaitr.geoPyramid = &this.pyramid.pyramids[v]
				minLevel = k
			}
		}
	}

	return feaitr
}

// 清空内存数据
func (this *MemFeaset) Close() {
	this.features = this.features[:0]

	// 清空索引
	if this.index != nil {
		this.index.Clear()
	}

	// 清空金字塔
	if this.pyramid != nil {
		this.pyramid.Clear()
	}
}

// 批量写入数据 todo
func (this *MemFeaset) BatchWrite(feas []Feature) {
}

func (this *MemFeaset) EndWrite() {
}

// 构建空间索引
func (this *MemFeaset) BuildSpatialIndex(indexType index.SpatialIndexType) index.SpatialIndex {
	if this.index == nil {
		tr := base.NewTimeRecorder()

		this.index = index.NewSpatialIndex(indexType)
		this.index.Init(this.bbox, int64(len(this.features)))
		for _, v := range this.features {
			this.index.AddGeo(v.Geo)
		}
		check := this.index.Check()
		fmt.Println("check building spatial index, result is:", check)

		tr.Output("build spatial index")
		return this.index
	}
	return nil
}

type MemFeaItr struct {
	ids              []int64              // id数组
	feaset           *MemFeaset           // 数据集指针
	geoPyramid       *[]geometry.Geometry // 金字塔层的对象
	objCountPerBatch int                  // 每个批次要读取的对象数量
	fields           []string             // 字段名，空则为所有字段
	squery           SpatailQuery
}

func (this *MemFeaItr) Count() int64 {
	return int64(len(this.ids))
}

func (this *MemFeaItr) Close() {
	this.ids = this.ids[:0]
	this.fields = this.fields[:0]
	this.feaset = nil
}

// todo
func (this *MemFeaItr) Next() (Feature, bool) {
	// if this.pos < int64(len(this.ids)) {
	// 	oldpos := this.pos
	// 	this.pos++
	// 	fea := this.feaset.features[this.ids[oldpos]]
	// 	return this.getFeaByAtts(fea), true
	// } else {
	// 	return *new(Feature), false
	// }
	return *new(Feature), false
}

// 为了批量读取做准备，返回批量的次数
func (this *MemFeaItr) PrepareBatch(objCount int) int {
	goCount := len(this.ids)/objCount + 1
	this.objCountPerBatch = objCount
	return goCount
}

// 批量读取支持go协程安全
func (this *MemFeaItr) BatchNext(batchNo int) (feas []Feature, result bool) {
	remainCount := len(this.ids) - batchNo*this.objCountPerBatch
	if remainCount >= 0 {
		objCount := this.objCountPerBatch
		if remainCount < objCount {
			objCount = remainCount
		}
		start := batchNo * this.objCountPerBatch
		feas = this.getFeaturesByIds(this.ids[start : start+objCount])
		result = true
	}
	return
}

func (this *MemFeaItr) getFeaturesByIds(ids []int64) []Feature {
	feas := make([]Feature, 0, len(ids))
	for _, id := range ids {
		if fea, ok := this.getOneFeature(id); ok {
			feas = append(feas, fea)
		}
	}
	return feas
}

// 返回 false，说明这个不能要
func (this *MemFeaItr) getOneFeature(id int64) (fea Feature, res bool) {
	if this.geoPyramid != nil {
		fea.Geo = (*this.geoPyramid)[id].Clone()
	} else {
		fea.Geo = this.feaset.features[id].Geo.Clone()
	}
	if this.squery.Match(fea.Geo) {
		this.setFeaAtts(fea, this.feaset.features[id])
		res = true
	}

	return
}

// 根据需要，只取一部分字段值
func (this *MemFeaItr) setFeaAtts(out, fea Feature) {
	out.Atts = make(map[string]interface{})
	// 所有属性全要
	if this.fields == nil || len(this.fields) == 0 {
		for k, v := range fea.Atts {
			out.Atts[k] = v
		}
	} else {
		// 根据 fields 来设置属性
		for _, field := range this.fields {
			out.Atts[field] = fea.Atts[field]
		}
	}
}
