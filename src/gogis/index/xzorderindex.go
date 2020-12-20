// XZ-Order 索引
package index

import (
	"gogis/base"
	"gogis/geometry"
	"io"
	"math"
	"sort"
)

// 控制每个最小的cell中，平均对象个数
const ONE_CELL_COUNT_X = 1000

type codeids struct {
	code int64
	ids  []int64
}

// todo 尚未做好，请勿使用
// 详情请参考相关论文
// 这里实现时，采用了自行设计的编码，主要是发挥计算机位运算的便利性
type XzorderIndex struct {
	bbox     base.Rect2D
	w, h     float64           // 最小的cell，宽和高
	level    int               // 划分的层级，0为整体，+1则一分为四
	code2ids map[int64][]int64 // code -->ids map格式，构建时使用
	// isOrder bool // 是否已经排序，当输入 geo个数等于 num时调用
	num, addedNum int64 // 判断是否需要对codes进行排序，以便后续的查询使用

	code2ids2 []codeids // code --> ids，查询时用这个
}

// 初始化
func (this *XzorderIndex) Init(bbox base.Rect2D, num int64) {
	this.bbox = bbox
	this.level = int(math.Log(float64(num)/ONE_CELL_COUNT_X)/2) + 1
	// 最多31层，int64最多存储32层，还要留一个bit做前置
	this.level = base.IntMin(this.level, 31)
	// 计算 一个轴方向的cell count
	oneAxisCount := math.Pow(2.0, float64(this.level))
	this.w = bbox.Dx() / oneAxisCount
	this.h = bbox.Dy() / oneAxisCount
	this.code2ids = make(map[int64][]int64, calcCellCount(int32(this.level)))
}

// 根据层级计算所有的cell个数
func calcCellCount(level int32) (count int) {
	for level >= 0 {
		count += int(math.Pow(4.0, float64(level)))
		level--
	}
	return
}

// 输入几何对象，构建索引；下列三种方式等效，同一个对象请勿重复调用Add方法
func (this *XzorderIndex) AddGeos(geometrys []geometry.Geometry) {
	for _, v := range geometrys {
		this.AddOne(v.GetBounds(), v.GetID())
	}
}

func (this *XzorderIndex) AddGeo(geo geometry.Geometry) {
	this.AddOne(geo.GetBounds(), geo.GetID())
}

func (this *XzorderIndex) AddOne(bbox base.Rect2D, id int64) {
	// 先根据bbox的大小和位置，确定层次
	level := this.ensureLevel(bbox)

	// 再计算bbox的code
	code := this.calcCode(bbox, level)

	_, ok := this.code2ids[code]
	if !ok {
		this.code2ids[code] = make([]int64, 0)
	}
	this.code2ids[code] = append(this.code2ids[code], id)

	this.addedNum++
	if this.addedNum >= this.num {
		this.adjustCode2ids()
	}
}

// 把map结构换做数组结构，以便后续查询使用
func (this *XzorderIndex) adjustCode2ids() {
	this.code2ids2 = make([]codeids, 0, len(this.code2ids))
	codes := make([]int64, 0, len(this.code2ids))
	for k, _ := range this.code2ids {
		codes = append(codes, k)
	}
	sort.Sort(base.Int64s(codes))

	for _, v := range codes {
		var one codeids
		one.code = v
		one.ids = this.code2ids[v]
		this.code2ids2 = append(this.code2ids2, one)
	}
}

// 确定最小合适的cell所在层级
func (this *XzorderIndex) ensureLevel(bbox base.Rect2D) (level int) {
	// 两个方向都试试看，哪个小用哪个
	xlevel := calcLevel(this.bbox.Min.X, this.w, bbox.Min.X, bbox.Dx(), this.level)
	ylevel := calcLevel(this.bbox.Min.Y, this.h, bbox.Min.Y, bbox.Dy(), this.level)
	return base.IntMin(xlevel, ylevel)
}

// 单个方向上计算层级
func calcLevel(anchor, cellLength, pos, length float64, level int) int {
	// 先看长度，大于cell的两倍长度，那当前的level肯定hold不住
	for length > cellLength*2 {
		cellLength *= 2
		level--
	}

	// 再看是否穿越了两个 cell，若是，层级要 -1
	pos -= anchor
	if pos+length > 2*cellLength {
		level--
	}

	return level
}

// 两个数组交叉合并
func combineBits(bits1 []byte, bits2 []byte) (bits []byte) {
	bits = make([]byte, 0, len(bits1)*2)
	for i, v := range bits {
		bits = append(bits, v)
		bits = append(bits, bits2[i])
	}
	return
}

// 已知层级，计算z编码
func (this *XzorderIndex) calcCode(bbox base.Rect2D, level int) (code int64) {
	xbits := calcBits(this.bbox.Min.X, this.w/2, bbox.Min.X, level)
	ybits := calcBits(this.bbox.Min.Y, this.h/2, bbox.Min.Y, level)
	bits := combineBits(xbits, ybits)

	code = 1 // 1 作为前置
	code = code << len(bits)
	for i := 0; i < len(bits); i++ {
		code &= (int64(bits[i]) << i)
	}
	return
}

// 计算一个方向的编码
func calcBits(anchor, halfLength, pos float64, level int) (bits []byte) {
	for level > 0 {
		// 小为0，大为1
		if pos < anchor+halfLength {
			bits = append(bits, 0)
		} else {
			bits = append(bits, 1)
			anchor += halfLength
		}
		level--
		halfLength /= 2
	}
	return
}

// 构建后，检查是否有问题；没问题返回true
func (this *XzorderIndex) Check() bool {
	return true
}

// 保存和加载，避免每次都要重复构建
func (this *XzorderIndex) Save(w io.Writer) {

}

func (this *XzorderIndex) Load(r io.Reader) {

}

// 范围查询，返回id数组
func (this *XzorderIndex) Query(bbox base.Rect2D) []int64 {
	// level := this.ensureLevel(bbox)
	// code := this.calcCode(bbox, level)

	return nil
}

// 清空
func (this *XzorderIndex) Clear() {

}

// 返回自己的类型
func (this *XzorderIndex) Type() SpatialIndexType {
	return TypeXzorderIndex
}
