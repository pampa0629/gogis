package index

import (
	"encoding/binary"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"io"
	"os"
)

// Add 协程不安全，请顺序调用
// Query方法 协程安全，可在并发调用
type SpatialIndex interface {
	// 初始化
	Init(bbox base.Rect2D, num int64)

	// 输入几何对象，构建索引；下列三种方式等效，同一个对象请勿重复调用Add方法
	AddGeos(geometrys []geometry.Geometry)
	AddGeo(geo geometry.Geometry)
	AddOne(bbox base.Rect2D, id int64) // 直接输入id和bbox

	// 构建后，检查是否有问题；没问题返回true
	Check() bool

	// 保存和加载，避免每次都要重复构建
	Save(w io.Writer)
	Load(r io.Reader)

	// 范围查询，返回id数组
	Query(bbox base.Rect2D) []int64

	// 查询不被bbox所覆盖的id数组
	QueryNoCovered(bbox base.Rect2D) []int64

	// 清空
	Clear()

	// 返回自己的类型
	Type() SpatialIndexType
}

// 几何对象外部存储方式定义
type SpatialIndexType int32

const (
	TypeNoIndex      SpatialIndexType = 0
	TypeGridIndex    SpatialIndexType = 1 // 格网索引；构建快，查询慢，结果不精确
	TypeQTreeIndex   SpatialIndexType = 2 // 四叉树索引；构建速度中等，查询速度快，结果精确
	TypeRTreeIndex   SpatialIndexType = 3 // R树索引；构建速度慢，查询速度快，结果精确
	TypeZOrderIndex  SpatialIndexType = 4 // Z-Order索引，通过生成空间key来查找对象，适合数据库的并发/分布式查询与读写
	TypeXzorderIndex SpatialIndexType = 5 // todo XZ-Order索引，通过生成空间key来查找对象，适合数据库的并发/分布式查询与读写
)

// 根据类型，创建空间索引对象
// 注意，这里并未构建好索引，需要调用者输入bbox和几何对象，完成构建过程
func NewSpatialIndex(indexType SpatialIndexType) SpatialIndex {
	var index SpatialIndex
	switch indexType {
	case TypeGridIndex:
		index = new(GridIndex)
	case TypeQTreeIndex:
		index = new(QTreeIndex)
	case TypeRTreeIndex:
		index = new(RTreeIndex)
	case TypeZOrderIndex:
		index = new(ZOrderIndex)
	// case TypeXzorderIndex:
	// 	index = new(XzorderIndex)
	default:
	}
	return index
}

// gogis index file mark
// const GIX_MARK = "gix"

// 装载gogis索引文件，返回索引对象
func LoadGix(gixfile string) (index SpatialIndex) {
	gix, _ := os.Open(gixfile)
	defer gix.Close()
	// 这里应根据文件头，确定具体的索引类型
	gixMark := make([]byte, 4)
	binary.Read(gix, binary.LittleEndian, gixMark)
	if string(gixMark[:3]) == base.EXT_SPATIAL_INDEX_FILE {
		var indexType int32
		binary.Read(gix, binary.LittleEndian, &indexType)
		index = NewSpatialIndex(SpatialIndexType(indexType))
		index.Load(gix)
		if !index.Check() {
			fmt.Println("index check is false!")
		}
	}
	return
}

// 保存空间索引到文件
func SaveGix(gixfile string, index SpatialIndex) {
	gix, _ := os.Create(gixfile)
	defer gix.Close()

	gix.WriteString(base.EXT_SPATIAL_INDEX_FILE + " ") // 加一个空格，凑四个字符
	binary.Write(gix, binary.LittleEndian, index.Type())
	index.Save(gix)
}

// 适合数据库使用的空间索引
// 与SpatialIndex的区别在于：
// *）初始化时，需要bbox和level（划分多少层级，以便后续每个geo的bbox计算得到code
// *）DB索引计算bbox对应的code，以便和geo一起存储到数据库中
// *）查询时，返回bbox所对应codes，以便从数据库中用codes获取geos
type SpatialIndexDB interface {
	// 初始化
	InitDB(bbox base.Rect2D, level int32)

	GetCode(bbox base.Rect2D) int32

	// 构建后，检查是否有问题；没问题返回true
	Check() bool

	// 范围查询，返回 code数组
	QueryDB(bbox base.Rect2D) []int32
	// 查询不被bbox所覆盖的code数组
	QueryNoCoveredDB(bbox base.Rect2D) []int32

	// 清空
	Clear()

	// 返回自己的类型
	Type() SpatialIndexType
}

func NewSpatialIndexDB(indexType SpatialIndexType) SpatialIndexDB {
	var index SpatialIndexDB
	switch indexType {
	// case TypeGridIndex:
	// 	index = new(GridIndex)
	// case TypeQTreeIndex:
	// 	index = new(QTreeIndex)
	// case TypeRTreeIndex:
	// 	index = new(RTreeIndex)
	case TypeZOrderIndex:
		index = new(ZOrderIndex)
	case TypeXzorderIndex:
		index = new(XzorderIndex)
	default:
	}
	return index
}
