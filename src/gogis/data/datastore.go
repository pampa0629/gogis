package data

import (
	"errors"
	"gogis/base"
	"gogis/geometry"
	"strings"
)

// func init() {
// 	// fmt.Println("init function --->")
// }

// 打开数据的连接参数
// 参数有哪些，根据具体store类型而定
type ConnParams map[string]interface{}

func NewConnParams() ConnParams {
	// map 必须要make一下才能用
	return make(map[string]interface{})
}

// 数据存储类型定义
type StoreType string

const (
	StoreShape       StoreType = "Shape"
	StoreShapeMemory StoreType = "ShapeMemory" // 内存模式的shape存储库
	StoreMemory      StoreType = "Memory"      // 纯内存模式
	StoreSqlite      StoreType = "Sqlite"
	StoreHbase       StoreType = "Hbase" // hbase
	StoreES          StoreType = "es"    // elasticsearch
)

type NewDatastoreFunc func() Datastore

var gNewDatastores map[StoreType]NewDatastoreFunc

// 支持用户自定义数据存储库类型
func RegisterDatastore(storyType StoreType, newfunc NewDatastoreFunc) {
	if gNewDatastores == nil {
		gNewDatastores = make(map[StoreType]NewDatastoreFunc)
	}
	gNewDatastores[storyType] = newfunc
}

func NewDatastore(storyType StoreType) Datastore {
	newfunc, ok := gNewDatastores[storyType]
	if ok {
		return newfunc()
	}
	return nil
}

// func NewDatastore(storyType StoreType) Datastore {
// 	switch storyType {
// 	case StoreShape:
// 		return new(ShapeStore)
// 	case StoreShapeMemory:
// 		return new(ShpmemStore)
// 	case StoreMemory:
// 		return new(MemoryStore)
// 	case StoreSqlite:
// 		return new(SqliteStore)
// 	case StoreHbase:
// 		return new(HbaseStore)
// 	case StoreES:
// 		return new(EsStore)
// 	}
// 	return nil
// }

// 数据存储库
type Datastore interface {
	Open(params ConnParams) (bool, error)
	GetType() StoreType // 得到存储类型
	GetConnParams() ConnParams

	GetFeasetByNum(num int) (Featureset, error)
	GetFeasetByName(name string) (Featureset, error)
	GetFeasetNames() []string

	CreateFeaset(name string, bbox base.Rect2D, geotype geometry.GeoType) Featureset

	Close() // 关闭，释放资源
}

// 数据集集合的数组，减少各个引擎的重复代码
type Feasets struct {
	feasets []Featureset
}

func (this *Feasets) GetFeasetByNum(num int) (Featureset, error) {
	if num >= 0 && num < len(this.feasets) {
		return this.feasets[num], nil
	}
	return nil, errors.New("num must big than zero and less the count of feature sets.")
}

func (this *Feasets) GetFeasetByName(name string) (Featureset, error) {
	for _, v := range this.feasets {
		if strings.ToUpper(v.GetName()) == strings.ToUpper(name) {
			return v, nil
		}
	}
	return nil, errors.New("cannot find the feature set of name: " + name + ".")
}

func (this *Feasets) GetFeasetNames() (names []string) {
	names = make([]string, len(this.feasets))
	for i, _ := range names {
		names[i] = this.feasets[i].GetName()
	}
	return
}

// 矢量数据集合
type Featureset interface {
	Open() (bool, error)
	Close()

	GetStore() Datastore
	GetName() string
	GetCount() int64 // 对象个数
	GetBounds() base.Rect2D
	GetGeoType() geometry.GeoType
	GetProjection() *base.ProjInfo // 得到投影坐标系，没有返回nil
	// GetFieldInfos() []FieldInfo

	// Query(bbox base.Rect2D, def QueryDef) FeatureIterator
	QueryByBounds(bbox base.Rect2D) FeatureIterator
	// QueryByDef(def QueryDef) FeatureIterator

	// 批量写入数据
	BatchWrite(feas []Feature)
	EndWrite()
}

// 辅助类，避免每个datastore都必须构造投影信息
type projCommon struct {
	proj *base.ProjInfo
}

// 得到投影坐标系，没有返回nil
func (this *projCommon) GetProjection() *base.ProjInfo {
	return this.proj
}

// 集合对象迭代器，用来遍历对象
type FeatureIterator interface {
	Count() int64
	// todo
	// Next() (Feature, bool)

	// 为调用批量读取做准备，调用 BatchNext 之前必须调用 本函数
	// objCount 为每个批次拟获取对象的数量，不保证精确
	PrepareBatch(objCount int) int

	// 批量读取，支持go协程安全；调用前，务必调用 PrepareBatch
	// batchNo 为批量的序号
	// 只要读取到一个数据，达不到count的要求，也返回true
	BatchNext(batchNo int) ([]Feature, bool)
	Close() // 关闭，释放资源
}

// 矢量要素（带属性值）
type Feature struct {
	Geo geometry.Geometry
	// todo which?
	// Fields map[string]interface{}
	Atts map[string]interface{}
	// Atts []string
}

func (this *Feature) Clone() Feature {
	var fea Feature
	fea.Geo = this.Geo.Clone()
	fea.Atts = make(map[string]interface{})
	for k, v := range this.Atts {
		fea.Atts[k] = v
	}
	return fea
}

// 栅格数据集合 todo
type Rasterset interface {
	GetBounds() base.Rect2D
}
