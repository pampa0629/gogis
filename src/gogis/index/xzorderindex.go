// XZ-Order 索引
package index

import (
	"bytes"
	"gogis/base"
	"math"
	"sort"
)

// 控制每个最小的cell中，平均对象个数
const ONE_CELL_COUNT_X = 1000

// todo 当前仅支持for数据库的索引接口
// 性能和ZOrder基本相当，但显示效果好（不会首先出现很远处与格子边界相交的对象）
// 详情请参考相关论文
type XzorderIndex struct {
	ZOrderIndex
}

func (this *XzorderIndex) InitDB(bbox base.Rect2D, level int32) {
	this.ZOrderIndex.InitDB(bbox, level)
}

func (this *XzorderIndex) GetCode(bbox base.Rect2D) int32 {
	bits := this.calcBboxBits(bbox)
	return Bits2code(bits)
}

// 本函数和ZOrder索引不同，由于XZ-Order的扩展能力，故而编码倾向于往min点
func (this *XzorderIndex) calcBboxBits(bbox base.Rect2D) (bits []byte) {
	// 思路：先分别计算 min和max两个点的code
	for level := this.level; level >= 0; level-- {
		minxBits := calcOneBits(this.bbox.Min.X, this.bbox.Dx()/2, bbox.Min.X, level, true)
		minyBits := calcOneBits(this.bbox.Min.Y, this.bbox.Dy()/2, bbox.Min.Y, level, true)

		maxxBits := calcOneBits(this.bbox.Min.X, this.bbox.Dx()/2, bbox.Max.X, level, false)
		maxyBits := calcOneBits(this.bbox.Min.Y, this.bbox.Dy()/2, bbox.Max.Y, level, false)

		// 再看max和min是否在一个格子内，或者max-1后，是否和min在一起
		x, y := false, false // 默认不相等
		maxxBits_1, _ := bitsReduce(maxxBits)
		if bytes.Equal(minxBits, maxxBits) || bytes.Equal(minxBits, maxxBits_1) {
			x = true
		}
		maxyBits_1, _ := bitsReduce(maxyBits)
		if bytes.Equal(minyBits, maxyBits) || bytes.Equal(minyBits, maxyBits_1) {
			y = true
		}
		if x && y { // x和y两个方向都ok，说明这个bbox可以用本层的格子编码
			bits = combineBits(minxBits, minyBits)
			break
		}
	}
	return
}

// 二进制数组的值减一，成功ok返回true；
// 若都是0，则无法减一，则ok返回false
func bitsReduce(bits []byte) (res []byte, ok bool) {
	last := len(bits) - 1
	if last >= 0 {
		res = make([]byte, len(bits))
		copy(res, bits)
		if res[last] == 1 {
			res[last] = 0
			ok = true
		} else {
			bits_1, ok_1 := bitsReduce(bits[0:last])
			if ok_1 {
				ok = true
				res[last] = 1
				copy(res[0:last], bits_1)
			}
		}
	}
	return
}

// 查询得到 bbox 所涉及的code，返回code数组（已排序）
func (this *XzorderIndex) QueryDB(bbox base.Rect2D) (codes []int32) {
	// 查询时，考虑XZ索引的扩展性，必须查询：
	// 1）bbox所在cell（原始编码）；2）上级cell；3）1和2的左、下、左下三个cell；4）下级cell

	// 思路：
	// 先找高层的level的code
	// 再迭代处理本层和低层
	//   得到bbox的code，
	//   低层的level，则先判断是否bbox相交，再迭代查询，直到最底层的level为止

	bits := this.ZOrderIndex.calcBboxBits(bbox)

	// 更高层的level
	upbits := make([]byte, len(bits))
	copy(upbits, bits)
	for len(upbits) > 0 {
		upbits = upbits[0 : len(upbits)-2] // 去掉最后两个bit，即提升一个level
		codes = append(codes, bits2Codes(upbits)...)
	}

	// 这里查询本层和下层的
	codes = append(codes, this.queryThisDownDB(bbox, bits, false)...)

	// 还要去掉重复的code
	codes = base.RemoveRepeatInt32(codes)

	// 最后排序
	sort.Sort(base.Int32s(codes))
	return
}

// 得到相关的几个编码
func bits2Codes(bits []byte) (codes []int32) {
	bitss := exBits(bits) // 考虑XZ-Order的扩展性，左、下、左下的几个格子也都要
	for _, v := range bitss {
		code := Bits2code(v)
		codes = append(codes, code)
	}
	return
}

// 查询本层以及下层的，需要迭代执行，直到最底层
// cover: 是否必须Cover；false：intersece即可
func (this *XzorderIndex) queryThisDownDB(bbox base.Rect2D, bits []byte, cover bool) (codes []int32) {
	// 得到本层的 codes
	codes = append(codes, bits2Codes(bits)...)

	// 若存在更低层的level，则构造下层的bits，判断bbox是否相交，再做查询
	if len(bits)/2 < int(this.level) { // 不是最底层
		downBitss := buildDownBitss(bits)
		for _, downBits := range downBitss {
			downBbox := this.calcBbox(downBits)
			if (cover && bbox.IsCovers(downBbox)) || bbox.IsIntersects(downBbox) {
				downCode := this.queryThisDownDB(bbox, downBits, cover)
				codes = append(codes, downCode...)
			}
		}
	}

	return
}

// 在ZOrder基础上，扩展一倍即可
func (this *XzorderIndex) calcBbox(bits []byte) base.Rect2D {
	bbox := this.ZOrderIndex.calcBbox(bits)
	bbox.Max.X += bbox.Dx()
	bbox.Max.Y += bbox.Dy()
	return bbox
}

// 考虑XZ-Order的扩展性，左、下、左下的几个格子也都要
func exBits(bits []byte) (bitss [][]byte) {
	// 自己先要了
	bitss = append(bitss, bits)

	// 再分解bits为xbits和ybits
	xbits, ybits := splitBits(bits)
	// 两个方向都 - 1
	xbits_1, okX := bitsReduce(xbits)
	ybits_1, okY := bitsReduce(ybits)
	if okX {
		bitss = append(bitss, combineBits(xbits_1, ybits))
	}
	if okY {
		bitss = append(bitss, combineBits(xbits, ybits_1))
	}
	if okX && okY {
		bitss = append(bitss, combineBits(xbits_1, ybits_1))
	}
	return
}

// 分解bits为xbits和ybits
func splitBits(bits []byte) (xbits, ybits []byte) {
	for i := 0; i < len(bits); i += 2 {
		xbits = append(xbits, bits[i+0])
		ybits = append(ybits, bits[i+1])
	}
	return
}

// 根据层级计算所有的cell个数
func calcCellCount(level int32) (count int32) {
	for level >= 0 {
		count += int32(math.Pow(4.0, float64(level)))
		level--
	}
	return
}

// 构建后，检查是否有问题；没问题返回true
func (this *XzorderIndex) Check() bool {
	return true
}

// todo
// 查询不被bbox所覆盖的id数组
func (this *XzorderIndex) QueryNoCoveredDB(bbox base.Rect2D) []int32 {
	return nil
}

// 清空
func (this *XzorderIndex) Clear() {

}

// 返回自己的类型
func (this *XzorderIndex) Type() SpatialIndexType {
	return TypeXzorderIndex
}
