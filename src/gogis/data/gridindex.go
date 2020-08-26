package data

import (
	"gogis/base"
	"gogis/geometry"
	"math"
)

// 每个格网控制在多少条记录左右
const ONE_GRID_COUNT = 10000

type GridIndex struct {
	indexs   [][][]int       // 格子编号对应 的ids
	bboxs    [][]base.Rect2D // 格子编号对应的bbox
	row, col int             // 格子 的行列数量
	len      float64         // 格子边长
	min      base.Point2D    // 左下角（最小值）的点
}

// 根据范围和 对象 数量来确定grid数量和每个grid的 box范围
func (this *GridIndex) Init(bbox base.Rect2D, num int) {
	// 先根据 ONE_GRID_COUNT 计算 应该 分为多少个 grid
	count := num/ONE_GRID_COUNT + 1
	// 再确定 格子的行列数，步骤：1）算出每个 grid的面积；2）算出grid边长；3）算出 x/y方向的grid个数
	area := bbox.Area() / float64(count)
	this.len = math.Sqrt(area)
	this.col = (int)(math.Ceil(bbox.Dx() / this.len))
	this.row = (int)(math.Ceil(bbox.Dy() / this.len))
	this.min = bbox.Min
	// fmt.Println("grid index:", count,area,this.len,this.col,this.row)

	cap := (int)(math.Min(float64(num), ONE_GRID_COUNT/4)) // 预留空间大小
	this.indexs = make([][][]int, this.row)
	for i, _ := range this.indexs {
		this.indexs[i] = make([][]int, this.col)
		for j, _ := range this.indexs[i] {
			this.indexs[i][j] = make([]int, 0, cap)
		}
	}

	this.bboxs = make([][]base.Rect2D, this.row)
	// 计算每个 bbox的边框
	for i, _ := range this.bboxs {
		this.bboxs[i] = make([]base.Rect2D, this.col)
		for j, _ := range this.bboxs[i] {
			this.bboxs[i][j].Min.X = this.min.X + float64(j)*this.len
			this.bboxs[i][j].Max.X = this.bboxs[i][j].Min.X + this.len
			this.bboxs[i][j].Min.Y = this.min.Y + float64(i)*this.len
			this.bboxs[i][j].Max.Y = this.bboxs[i][j].Min.Y + this.len
		}
	}
}

func (this *GridIndex) Clear() {
	this.indexs = this.indexs[:0]
	this.bboxs = this.bboxs[:0]
}

// 构建空间索引
func (this *GridIndex) BuildByGeos(geometrys []geometry.Geometry) {
	for i, geo := range geometrys {
		this.dealOneGeo(geo, i)
	}
	// fmt.Println("grid indexes:", this.indexs)
}

func (this *GridIndex) BuildByFeas(features []Feature) {
	for i, fea := range features {
		this.dealOneGeo(fea.geo, i)
	}
	// fmt.Println("grid indexes:", this.indexs)
}

func (this *GridIndex) dealOneGeo(geo geometry.Geometry, id int) {
	minRow, maxRow, minCol, maxCol := this.GetGridNo(geo.GetBounds())

	// 最后赋值
	for i := minRow; i <= maxRow; i++ { // 高度（y方向）代表行
		for j := minCol; j <= maxCol; j++ {
			this.indexs[i][j] = append(this.indexs[i][j], id)
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
	maxCol = base.IntMin(maxCol, this.col-1) // 不能大于 col数
	maxRow = (int)(math.Floor((bbox.Max.Y - this.min.Y) / this.len))
	maxRow = base.IntMin(maxRow, this.row-1) // 不能大于 row数
	return
}
