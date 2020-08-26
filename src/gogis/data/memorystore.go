package data

import (
	"errors"
	"fmt"
	"gogis/base"
)

type MemoryStore struct {
	feasets []*MemFeaset
}

// todo
func (this *MemoryStore) Open(params ConnParams) (bool, error) {
	return true, nil
}

func (this *MemoryStore) GetFeatureset(name string) (Featureset, error) {
	for _, v := range this.feasets {
		if v.name == name {
			return v, nil
		}
	}
	return nil, errors.New("feature set: " + name + " cannot find")
}

func (this *MemoryStore) FeaturesetNames() []string {
	names := make([]string, len(this.feasets))
	for i, v := range this.feasets {
		names[i] = v.name
	}
	return names
}

// 关闭，释放资源
func (this *MemoryStore) Close() {
	for _, feaset := range this.feasets {
		feaset.Close()
	}
	this.feasets = this.feasets[:0]
}

type MemFeaItr struct {
	ids    []int      // id数组
	feaset *MemFeaset // 数据集指针
	pos    int        // 当前位置
}

func (this *MemFeaItr) Next() (Feature, bool) {
	if this.pos < len(this.ids) {
		oldpos := this.pos
		this.pos++
		return this.feaset.features[this.ids[oldpos]], true
	} else {
		return *new(Feature), false
	}
}

// 内存矢量数据集
type MemFeaset struct {
	name     string
	bbox     base.Rect2D
	features []Feature      // 几何对象的数组
	index    *GridIndex     // 空间索引
	pyramid  *VectorPyramid // 矢量金字塔
}

func (this *MemFeaset) Open(name string) (bool, error) {
	return true, nil
}

func (this *MemFeaset) GetName() string {
	return this.name
}

func (this *MemFeaset) GetBounds() base.Rect2D {
	return this.bbox
}

// 根据空间范围查询，返回范围内geo的ids
func (this *MemFeaset) Query(bbox base.Rect2D) FeatureIterator {
	var feaitr MemFeaItr
	feaitr.feaset = this
	feaitr.ids = make([]int, 0)
	minRow, maxRow, minCol, maxCol := this.index.GetGridNo(bbox)

	// 最后赋值
	for i := minRow; i <= maxRow; i++ { // 高度（y方向）代表行
		for j := minCol; j <= maxCol; j++ {
			feaitr.ids = append(feaitr.ids, this.index.indexs[i][j]...)
		}
	}

	// 这里应该还要去掉重复id todo ......
	return &feaitr
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

// 构建空间索引
func (this *MemFeaset) BuildSpatialIndex() {
	if this.index == nil {
		this.index = new(GridIndex)
		this.index.Init(this.bbox, len(this.features))
		this.index.BuildByFeas(this.features)
	}
}

// 计算索引重复度，为后续有可能增加多级格网做准备
func (this *MemFeaset) calcRepeatability() float64 {
	indexCount := 0.0
	for i := 0; i < this.index.row; i++ {
		for j := 0; j < this.index.col; j++ {
			indexCount += float64(len(this.index.indexs[i][j]))
		}
	}
	repeat := indexCount / float64(len(this.features))
	fmt.Println("shp index重复度为:", repeat)
	return repeat
}

func (this *MemFeaset) BuildPyramids() {
	this.pyramid = new(VectorPyramid)
	this.pyramid.Build(this.bbox, this.features)
}
