package data

import (
	"errors"
	"gogis/base"
	"gogis/geometry"
	"strconv"
	"strings"
	"time"
)

func init() {
	// fmt.Println("init function --->")
}

// 打开数据的连接参数
// 参数有哪些，根据具体store类型而定
type ConnParams map[string]string

func NewConnParams() ConnParams {
	// map 必须要make一下才能用
	return make(map[string]string)
}

// 数据存储类型定义
type StoreType string

const (
	StoreShape       StoreType = "Shape"
	StoreShapeMemory StoreType = "ShapeMemory" // 内存模式的shape存储库
	StoreMemory      StoreType = "Memory"
	StoreSqlite      StoreType = "Sqlite"
)

func NewDatastore(storyType StoreType) Datastore {
	switch storyType {
	case StoreShape:
		return new(ShapeStore)
	case StoreShapeMemory:
		return new(ShpmemStore)
	case StoreMemory:
		return new(MemoryStore)
	case StoreSqlite:
		return new(SqliteStore)
	}
	return nil
}

// 数据存储库
type Datastore interface {
	Open(params ConnParams) (bool, error)
	GetType() StoreType // 得到存储类型
	GetConnParams() ConnParams

	GetFeasetByNum(num int) (Featureset, error)
	GetFeasetByName(name string) (Featureset, error)
	FeaturesetNames() []string

	Close() // 关闭，释放资源
}

// 矢量数据集合
type Featureset interface {
	Open(name string) (bool, error)
	Close()

	GetStore() Datastore
	GetName() string
	Count() int64 // 对象个数
	GetBounds() base.Rect2D
	GetFieldInfos() []FieldInfo

	Query(bbox base.Rect2D, def QueryDef) FeatureIterator
	QueryByBounds(bbox base.Rect2D) FeatureIterator
	QueryByDef(def QueryDef) FeatureIterator
}

// type FieldCompOp int

// const (
// 	UnknownOperator FieldCompOp = iota // UnknownOperator is the zero value for an Operator
// 	Eq                                 // Eq -> "="
// 	Ne                                 // Ne -> "!="
// 	Gt                                 // Gt -> ">"
// 	Lt                                 // Lt -> "<"
// 	Gte                                // Gte -> ">="
// 	Lte                                // Lte -> "<="
// )

// 属性查询条件定义
type QueryDef struct {
	Fields []string // 需要哪些字段
	Wheres []string // {Field1="abc",Field2>=10,......}
}

// 字段比较
type FieldComp struct {
	Field string
	Op    string
	// Value string
	Value interface{}
}

// 内部使用
func splitByMoreStr(r rune) bool {
	// 仅支持字段比较用的分隔符
	return r == '=' || r == '>' || r == '<' || r == '!'
}

// 内部解析wheres条件
func (this *QueryDef) Parser(finfos []FieldInfo) (comps []FieldComp, err error) {
	for _, where := range this.Wheres {
		res := strings.FieldsFunc(where, splitByMoreStr)
		if len(res) == 2 {
			op := strings.Trim(where, res[0])
			op = strings.Trim(op, res[1])
			op = strings.TrimSpace(op)
			ftype := GetFieldTypeByName(finfos, res[0])
			if ftype != TypeUnknown {
				value := string2value(res[1], ftype)
				newComp := FieldComp{res[0], op, value}
				// fmt.Println(newComp)
				comps = append(comps, newComp)
			} else {
				err = errors.New(res[0] + "'s field type is unknown.")
			}
		} else {
			err = errors.New(where + " cannot be parsed.")
		}
	}
	return
}

func GetFieldTypeByName(finfos []FieldInfo, name string) FieldType {
	for _, finfo := range finfos {
		if strings.ToUpper(finfo.Name) == strings.ToUpper(name) {
			return finfo.Type
		}
	}
	return TypeUnknown
}

// 把字符串性质的值，根据字段类型，转化为特定数据类型
func string2value(str string, ftype FieldType) interface{} {
	switch ftype {
	case TypeBool:
		// case "1", "t", "T", "true", "TRUE", "True":
		// case "0", "f", "F", "false", "FALSE", "False":
		value, _ := strconv.ParseBool(str)
		return value
	case TypeInt:
		value, _ := strconv.Atoi(str)
		return value
	case TypeFloat:
		value, _ := strconv.ParseFloat(str, 64)
		return value
	case TypeString:
		return str
	case TypeTime:
		value, _ := time.Parse(TIME_LAYOUT, str)
		return value
	case TypeBlob:
		return []byte(str)
	}
	return nil
}

type FieldType int32

const (
	TypeUnknown FieldType = 0
	TypeBool    FieldType = 1 // bool
	TypeInt     FieldType = 2 // int32
	TypeFloat   FieldType = 3 // float64
	TypeString  FieldType = 5 // string
	TypeTime    FieldType = 6 // time.Time
	TypeBlob    FieldType = 7 // []byte
)

// 字段描述信息
type FieldInfo struct {
	Name   string
	Type   FieldType
	Length int
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

// 一个矢量对象（带属性）
type Feature struct {
	Geo geometry.Geometry
	// todo which?
	// Fields map[string]interface{}
	Atts map[string]interface{}
	// Atts []string
}

// 栅格数据集合 todo
type Rasterset interface {
	GetBounds() base.Rect2D
}
