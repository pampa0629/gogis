package data

import (
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"gogis/index"
	"time"
)

type MemoryStore struct {
	// feasets []*MemFeaset
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

// func (this *MemoryStore) GetFeasetByNum(num int) (Featureset, error) {
// 	if num < len(this.feasets) {
// 		return this.feasets[num], nil
// 	}
// 	return nil, errors.New(strconv.Itoa(num) + " beyond the num of feature sets")
// }

// func (this *MemoryStore) GetFeasetByName(name string) (Featureset, error) {
// 	for _, v := range this.feasets {
// 		if v.name == name {
// 			return v, nil
// 		}
// 	}
// 	return nil, errors.New("feature set: " + name + " cannot find")
// }

// func (this *MemoryStore) GetFeasetNames() []string {
// 	names := make([]string, len(this.feasets))
// 	for i, v := range this.feasets {
// 		names[i] = v.name
// 	}
// 	return names
// }

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

// 范围和属性联合查询 todo
// func (this *MemFeaset) Query(bbox base.Rect2D, def QueryDef) FeatureIterator {
// 	return nil
// }

// 属性查询
func (this *MemFeaset) QueryByDef(def QueryDef) FeatureIterator {
	var feaitr MemFeaItr
	feaitr.feaset = this
	feaitr.fields = def.Fields
	feaitr.ids = make([]int64, 0)

	// 先解析 wheres语句
	comps, _ := def.Parser(this.fieldInfos)
	// 得到需要 处理的字段的类型
	ftypes := make([]FieldType, len(comps))
	for i, comp := range comps {
		ftypes[i] = GetFieldTypeByName(this.fieldInfos, comp.Field)
	}

	for i, fea := range this.features {
		if IsAllMatch(fea, comps, ftypes) {
			feaitr.ids = append(feaitr.ids, int64(i))
		}
	}

	return &feaitr
}

// 判断feature是否符合属性要求
func IsAllMatch(fea Feature, comps []FieldComp, ftypes []FieldType) bool {
	// 有一条不符合，就返回false
	for i, comp := range comps {
		switch ftypes[i] {
		case TypeBool:
			if !base.IsMatchBool(fea.Atts[comp.Field].(bool), comp.Op, comp.Value.(bool)) {
				return false
			}
		case TypeInt:
			if !base.IsMatchInt(fea.Atts[comp.Field].(int), comp.Op, comp.Value.(int)) {
				return false
			}
		case TypeFloat:
			if !base.IsMatchFloat(fea.Atts[comp.Field].(float64), comp.Op, comp.Value.(float64)) {
				return false
			}
		case TypeString:
			if !base.IsMatchString(fea.Atts[comp.Field].(string), comp.Op, comp.Value.(string)) {
				return false
			}
		case TypeTime:
			if !base.IsMatchTime(fea.Atts[comp.Field].(time.Time), comp.Op, comp.Value.(time.Time)) {
				return false
			}
		case TypeBlob:
			// 暂不支持
		}
	}
	// 每一条都符合，才能通过
	return true
}

// func isMatchBool(value interface{}, op string, value2 string, ftype FieldType) bool {
// 	// 每一条都符合，才能通过

// 	return true
// }

// 根据空间范围查询，返回范围内geo的ids
func (this *MemFeaset) QueryByBounds(bbox base.Rect2D) FeatureIterator {
	feaitr := new(MemFeaItr)
	feaitr.feaset = this
	feaitr.ids = this.index.Query(bbox)
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

func (this *MemFeaset) BuildPyramids() {
	this.pyramid = new(VectorPyramid)
	this.pyramid.Build(this.bbox, this.features)
}

type MemFeaItr struct {
	ids              []int64    // id数组
	feaset           *MemFeaset // 数据集指针
	objCountPerBatch int        // 每个批次要读取的对象数量
	fields           []string   // 字段名，空则为所有字段
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

// 根据需要，只取一部分字段值
func (this *MemFeaItr) getFeaByAtts(fea Feature) Feature {
	if this.fields == nil || len(this.fields) == 0 {
		return fea
	}
	newfea := new(Feature)
	newfea.Geo = fea.Geo
	newfea.Atts = make(map[string]interface{})
	for _, field := range this.fields {
		newfea.Atts[field] = fea.Atts[field]
	}
	return *newfea
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
		feas = this.getFeasByIds(this.feaset.features, this.ids[start:start+objCount])
		result = true
	}
	return
}

// 批量读取支持go协程安全
// func (this *MemFeaItr) BatchNext2(pos int64, count int) ([]Feature, int64, bool) {
// 	len := int64(len(this.ids))
// 	if pos < len {
// 		oldpos := pos
// 		if int64(count)+pos > len {
// 			count = int(len - pos)
// 		}
// 		pos += int64(count)
// 		return this.getFeasByIds(this.feaset.features, this.ids[oldpos:oldpos+int64(count)]), pos, true
// 	} else {
// 		return nil, pos, false
// 	}
// }

// func (this *MemFeaItr) BatchNext2(pos int, count int) ([]Feature, int, bool) {
// 	len := int64(len(this.ids))
// 	if this.pos < len {
// 		oldpos := this.pos
// 		if int64(count)+this.pos > len {
// 			count = int(len - this.pos)
// 		}
// 		this.pos += int64(count)
// 		return this.getFeasByIds(this.feaset.features, this.ids[oldpos:oldpos+int64(count)]), true
// 	} else {
// 		return nil, false
// 	}
// }

func (this *MemFeaItr) getFeasByIds(features []Feature, ids []int64) []Feature {
	newfeas := make([]Feature, len(ids))
	for i, id := range ids {
		// pos := this.feaset.id2feaPos[id]
		// newfeas[i] = this.getFeaByAtts(features[pos])
		newfeas[i] = this.getFeaByAtts(features[id])
	}
	return newfeas
}
