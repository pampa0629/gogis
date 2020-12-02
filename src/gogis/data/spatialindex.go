package data

import (
	"encoding/binary"
	"gogis/base"
	"gogis/geometry"
	"io"
	"os"
)

type SpatialIndex interface {
	// 初始化
	Init(bbox base.Rect2D, num int64)

	// 输入几何对象，构建索引
	BuildByGeos(geometrys []geometry.Geometry)
	BuildByFeas(features []Feature)

	// 构建后，检查是否有问题；没问题返回true
	Check() bool

	// 保存和加载，避免每次都要重复构建
	Save(w io.Writer)
	Load(r io.Reader)

	// 范围查询，返回id数组
	Query(bbox base.Rect2D) []int64

	// 清空
	Clear()

	// 返回自己的类型
	Type() SpatialIndexType
}

// 几何对象外部存储方式定义
type SpatialIndexType int32

const (
	TypeNoIndex    SpatialIndexType = 0
	TypeGridIndex  SpatialIndexType = 1 // 格网索引；构建快，查询慢，结果不精确
	TypeQTreeIndex SpatialIndexType = 2 // 四叉树索引；构建速度中等，查询速度快，结果精确
	TypeRTreeIndex SpatialIndexType = 3 // R树索引；构建速度慢，查询速度快，结果精确
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
	default:
	}
	return index
}

// gogis index file mark
const GIX_MARK = "gix"

// 装载gogis索引文件，返回索引对象
func loadGix(gixfile string) (index SpatialIndex) {
	gix, _ := os.Open(gixfile)
	defer gix.Close()
	// 这里应根据文件头，确定具体的索引类型
	gixMark := make([]byte, 4)
	binary.Read(gix, binary.LittleEndian, gixMark)
	if string(gixMark[:3]) == GIX_MARK {
		var indexType int32
		binary.Read(gix, binary.LittleEndian, &indexType)
		index = NewSpatialIndex(SpatialIndexType(indexType))
		index.Load(gix)
	}
	return
}

// 保存空间索引到文件
func saveGix(gixfile string, index SpatialIndex) {
	gix, _ := os.Create(gixfile)
	defer gix.Close()

	gix.WriteString(GIX_MARK + " ") // 加一个空格，凑四个字符
	binary.Write(gix, binary.LittleEndian, index.Type())
	index.Save(gix)
}
