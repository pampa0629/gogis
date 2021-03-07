package data

import (
	"gogis/base"
	"sync"

	// "gogis/data/memory"
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

func (this *ConnParams) GetString(key string) string {
	iv := (*this)[key]
	if iv != nil {
		if value, ok := iv.(string); ok {
			return value
		}
	}
	return ""
}

// 数据存储类型定义
type StoreType string

const (
	StoreUnknown StoreType = "Unknown" //
	StoreMemory  StoreType = "Memory"  // 纯内存模式
	StoreShape   StoreType = "Shape"
	StoreSqlite  StoreType = "Sqlite"
	StoreHbase   StoreType = "Hbase" // hbase
	StoreES      StoreType = "ES"    // elasticsearch
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

// 数据存储库
type Datastore interface {
	Open(params ConnParams) (bool, error)
	GetType() StoreType // 得到存储类型
	GetConnParams() ConnParams

	GetFeasetByNum(num int) Featureset
	GetFeasetByName(name string) (Featureset, int)
	GetFeasetNames() []string

	// CreateFeaset(name string, bbox base.Rect2D, geotype geometry.GeoType) Featureset
	CreateFeaset(info FeasetInfo) Featureset
	DeleteFeaset(name string) bool

	Close() // 关闭，释放资源
}

// ============================================================================ //

// 数据集集合的数组，减少各个引擎的重复代码
type Featuresets struct {
	Feasets []Featureset
}

func (this *Featuresets) GetFeasetByNum(num int) Featureset {
	if num >= 0 && num < len(this.Feasets) {
		return this.Feasets[num]
	}
	return nil // , errors.New("num must big than zero and less the count of feature sets.")
}

func (this *Featuresets) GetFeasetByName(name string) (Featureset, int) {
	for i, v := range this.Feasets {
		if strings.ToUpper(v.GetName()) == strings.ToUpper(name) {
			return v, i
		}
	}
	return nil, -1 // , errors.New("cannot find the feature set of name: " + name + ".")
}

func (this *Featuresets) GetFeasetNames() (names []string) {
	names = make([]string, len(this.Feasets))
	for i, _ := range names {
		names[i] = this.Feasets[i].GetName()
	}
	return
}

// ============================================================================ //

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
	GetFieldInfos() []FieldInfo

	// Query(bbox base.Rect2D, def QueryDef) FeatureIterator
	// QueryByBounds(bbox base.Rect2D) FeatureIterator
	// 根据指定条件查询；输入nil时，按照全幅范围、所有字段查询
	Query(def *QueryDef) FeatureIterator

	BeforeWrite(count int64)
	// 批量写入数据
	BatchWrite(feas []Feature)
	EndWrite()
}

// ============================================================================ //
const ONE_LOAD_COUNT = 100000

// 缓存到内存中
func Cache(in Featureset, fields []string) (out Featureset) {
	store := NewDatastore(StoreMemory)
	store.Open(in.GetStore().GetConnParams())
	var info FeasetInfo
	info.Bbox = in.GetBounds()
	info.Name = in.GetName()
	info.GeoType = in.GetGeoType()
	info.FieldInfos = in.GetFieldInfos()
	out = store.CreateFeaset(info)

	var def QueryDef
	def.Fields = fields
	def.SpatialObj = in.GetBounds()
	feait := in.Query(&def)
	forCount := feait.BeforeNext(ONE_LOAD_COUNT)
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < forCount; i++ {
		wg.Add(1)
		go func(feait FeatureIterator, n int, out Featureset, wg *sync.WaitGroup) {
			defer wg.Done()
			feas, ok := feait.BatchNext(n)
			if ok {
				out.BatchWrite(feas)
			}
		}(feait, i, out, wg)
	}
	wg.Wait()
	out.EndWrite()
	return out
}

// 要素数据集的基本信息
type FeasetInfo struct {
	Name       string
	Bbox       base.Rect2D
	GeoType    geometry.GeoType
	FieldInfos []FieldInfo
	Proj       *base.ProjInfo
}

func (this *FeasetInfo) GetGeoType() geometry.GeoType {
	return this.GeoType
}

func (this *FeasetInfo) GetName() string {
	return this.Name
}

func (this *FeasetInfo) GetBounds() base.Rect2D {
	return this.Bbox
}

func (this *FeasetInfo) GetFieldInfos() []FieldInfo {
	return this.FieldInfos
}

// 辅助类，避免每个datastore都必须构造投影信息
// type ProjCommon struct {

// }

// 得到投影坐标系，没有返回nil
func (this *FeasetInfo) GetProjection() *base.ProjInfo {
	return this.Proj
}

// ============================================================================ //

// 集合对象迭代器，用来遍历对象
type FeatureIterator interface {
	Count() int64
	// todo
	// Next() (Feature, bool)

	// 为调用批量读取做准备，调用 BatchNext 之前必须调用 本函数
	// objCount 为每个批次拟获取对象的数量，不保证精确
	BeforeNext(objCount int) int

	// 批量读取，支持go协程安全；调用前，务必调用 BeforeNext；
	// batchNo 为批量的序号；
	// 只要读取到一个数据，达不到count的要求，也返回true；若bool返回false，表示读取结束
	BatchNext(batchNo int) ([]Feature, bool)
	Close() // 关闭，释放资源
}

// ============================================================================ //
type Atts map[string]interface{}

// 矢量要素（带属性值）
type Feature struct {
	Geo geometry.Geometry
	// todo which?
	// Fields map[string]interface{}
	Atts
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

// ============================================================================ //

// 栅格数据集合 todo
type Rasterset interface {
	GetBounds() base.Rect2D
}
