// 格网索引
package data

import (
	"encoding/binary"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"io"
	"math"
)

// 每个格网控制在多少条记录左右
const ONE_GRID_COUNT = 10000

type GridIndex struct {
	ids      [][][]int64     // 格子编号对应 的ids
	bboxes   [][]base.Rect2D // 格子编号对应的bbox
	row, col int32           // 格子 的行列数量
	len      float64         // 格子边长
	min      base.Point2D    // 左下角（最小值）的点
}

// 保存和加载，避免每次都要重复构建
func (this *GridIndex) Save(w io.Writer) {
	binary.Write(w, binary.LittleEndian, this.min)
	binary.Write(w, binary.LittleEndian, this.len)
	binary.Write(w, binary.LittleEndian, this.row)
	binary.Write(w, binary.LittleEndian, this.col)
	binary.Write(w, binary.LittleEndian, this.bboxes)
	for i := int32(0); i < this.row; i++ {
		for j := int32(0); j < this.col; j++ {
			count := int32(len(this.ids[i][j]))
			binary.Write(w, binary.LittleEndian, count)
			binary.Write(w, binary.LittleEndian, this.ids[i][j])
		}
	}
}

func (this *GridIndex) Load(r io.Reader) {
	binary.Read(r, binary.LittleEndian, &this.min)
	binary.Read(r, binary.LittleEndian, &this.len)
	binary.Read(r, binary.LittleEndian, &this.row)
	binary.Read(r, binary.LittleEndian, &this.col)

	this.bboxes = make([][]base.Rect2D, this.row)
	for i := int32(0); i < this.row; i++ {
		this.bboxes[i] = make([]base.Rect2D, this.col)
	}
	binary.Read(r, binary.LittleEndian, this.bboxes)

	this.ids = make([][][]int64, this.row)
	for i := int32(0); i < this.row; i++ {
		this.ids[i] = make([][]int64, this.col)
		for j := int32(0); j < this.col; j++ {
			var count int32
			binary.Read(r, binary.LittleEndian, &count)
			this.ids[i][j] = make([]int64, count)
			binary.Read(r, binary.LittleEndian, this.ids[i][j])
		}
	}
}

// 构建后，检查是否有问题；没问题返回true
func (this *GridIndex) Check() bool {
	// 貌似也就只能检查bbox了
	for i := int32(0); i < this.row; i++ {
		for j := int32(0); j < this.col; j++ {
			if !base.IsEqual(this.bboxes[i][j].Min.X, this.min.X+float64(i)*this.len) {
				return false
			}
			if !base.IsEqual(this.bboxes[i][j].Min.Y, this.min.Y+float64(j)*this.len) {
				return false
			}
			if !base.IsEqual(this.bboxes[i][j].Min.X+this.len, this.bboxes[i][j].Max.X) {
				return false
			}
			if !base.IsEqual(this.bboxes[i][j].Min.Y+this.len, this.bboxes[i][j].Max.Y) {
				return false
			}
		}
	}
	// 再查查 len
	if this.row != int32(len(this.bboxes)) {
		return false
	}
	if this.col != int32(len(this.bboxes[0])) {
		return false
	}
	return true
}

func (this *GridIndex) Type() SpatialIndexType {
	return TypeGridIndex
}

// 根据范围和 对象 数量来确定grid数量和每个grid的 box范围
func (this *GridIndex) Init(bbox base.Rect2D, num int64) {
	// 先根据 ONE_GRID_COUNT 计算 应该 分为多少个 grid
	count := num/ONE_GRID_COUNT + 1
	// 再确定 格子的行列数，步骤：1）算出每个 grid的面积；2）算出grid边长；3）算出 x/y方向的grid个数
	area := bbox.Area() / float64(count)
	this.len = math.Sqrt(area)
	this.col = (int32)(math.Ceil(bbox.Dx() / this.len))
	this.row = (int32)(math.Ceil(bbox.Dy() / this.len))
	this.min = bbox.Min
	fmt.Println("grid index:", count, area, this.len, this.col, this.row)

	cap := (int)(math.Min(float64(num), ONE_GRID_COUNT/4)) // 预留空间大小
	this.ids = make([][][]int64, this.row)
	for i, _ := range this.ids {
		this.ids[i] = make([][]int64, this.col)
		for j, _ := range this.ids[i] {
			this.ids[i][j] = make([]int64, 0, cap)
		}
	}

	this.bboxes = make([][]base.Rect2D, this.row)
	// 计算每个 bbox的边框
	for i, _ := range this.bboxes {
		this.bboxes[i] = make([]base.Rect2D, this.col)
		for j, _ := range this.bboxes[i] {
			this.bboxes[i][j].Min.X = this.min.X + float64(j)*this.len
			this.bboxes[i][j].Max.X = this.bboxes[i][j].Min.X + this.len
			this.bboxes[i][j].Min.Y = this.min.Y + float64(i)*this.len
			this.bboxes[i][j].Max.Y = this.bboxes[i][j].Min.Y + this.len
		}
	}
}

func (this *GridIndex) Clear() {
	this.ids = this.ids[:0]
	this.bboxes = this.bboxes[:0]
}

// 构建空间索引
func (this *GridIndex) BuildByGeos(geometrys []geometry.Geometry) {
	for i, geo := range geometrys {
		this.dealOneGeo(geo, int64(i))
	}
	// fmt.Println("grid indexes:", this.ids)
}

func (this *GridIndex) BuildByFeas(features []Feature) {
	for i, fea := range features {
		this.dealOneGeo(fea.Geo, int64(i))
	}
	// fmt.Println("grid indexes:", this.ids)
}

func (this *GridIndex) dealOneGeo(geo geometry.Geometry, id int64) {
	minRow, maxRow, minCol, maxCol := this.GetGridNo(geo.GetBounds())

	// 最后赋值
	for i := minRow; i <= maxRow; i++ { // 高度（y方向）代表行
		for j := minCol; j <= maxCol; j++ {
			this.ids[i][j] = append(this.ids[i][j], id)
		}
	}
}

func (this *GridIndex) GetGridNo(bbox base.Rect2D) (minRow, maxRow, minCol, maxCol int) {
	// 先计算 起始点在哪个grid(行列号)
	minCol = (int)(math.Floor((bbox.Min.X - this.min.X) / this.len))
	minCol = base.IntMax(minCol, 0) // 不能小于0
	minRow = (int)(math.Floor((bbox.Min.Y - this.min.Y) / this.len))
	minRow = base.IntMax(minRow, 0) // 不能小于0

	// 再计算 终止点在哪个grid(行列号)
	maxCol = (int)(math.Floor((bbox.Max.X - this.min.X) / this.len))
	maxCol = base.IntMin(maxCol, int(this.col-1)) // 不能大于 col数
	maxRow = (int)(math.Floor((bbox.Max.Y - this.min.Y) / this.len))
	maxRow = base.IntMin(maxRow, int(this.row-1)) // 不能大于 row数
	return
}

func (this *GridIndex) Query(bbox base.Rect2D) (ids []int64) {
	minRow, maxRow, minCol, maxCol := this.GetGridNo(bbox)
	// 预估一下可能的ids容量
	cap := (maxRow - minRow) * (maxCol - minCol) * ONE_GRID_COUNT
	ids = make([]int64, 0, cap)

	// 最后赋值
	for i := minRow; i <= maxRow; i++ { // 高度（y方向）代表行
		for j := minCol; j <= maxCol; j++ {
			ids = append(ids, this.ids[i][j]...)
		}
	}

	// 去掉重复id
	ids = base.RemoveRepByMap(ids)
	return
}

// 计算索引重复度，为后续有可能增加多级格网做准备
func (this *GridIndex) calcRepeatability(count int) float64 {
	indexCount := 0.0
	for i := int32(0); i < this.row; i++ {
		for j := int32(0); j < this.col; j++ {
			indexCount += float64(len(this.ids[i][j]))
		}
	}
	repeat := indexCount / float64(count)
	fmt.Println("GridIndex重复度为:", repeat)
	return repeat
}
