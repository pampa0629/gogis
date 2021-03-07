// GeoHash 空间索引
package index

import (
	"gogis/base"
)

type GeoHash struct {
}

// 编码转坐标，返回编码格子中心点坐标，以及格子的宽和高
func (geohash GeoHash) GetCell(code string) (center base.Point2D, width, height float64) {
	// code --> bits
	codeLen := len(code)
	bits := make([]byte, 5*codeLen)
	for i := 0; i < codeLen; i++ {
		var b32 Base32
		b32 = Base32(code[i])
		copy(bits[i*5:i*5+5], b32.toBits())
	}

	// bits --> pos
	return hashBits2Cell(bits)
}

func hashBits2Cell(bits []byte) (center base.Point2D, width, height float64) {
	// 分解 bits 为 xbits和ybits
	var xbits, ybits []byte
	for i, bit := range bits {
		if i%2 == 0 {
			xbits = append(xbits, bit)
		} else {
			ybits = append(ybits, bit)
		}
	}

	// 分辨计算X和Y
	center.X, width = hashCalcOnePos(xbits, -180.0, 180.0)
	center.Y, height = hashCalcOnePos(ybits, -90.0, 90.0)
	return
}

// 根据bits，计算一个方向上的坐标位置，以及对应格子的长度
func hashCalcOnePos(bits []byte, min, max float64) (pos, length float64) {
	half := (max - min) / 2.0
	pos = (max + min) / 2.0 // 默认为中心点
	for i := 0; i < len(bits); i++ {
		// 减半再使用
		half /= 2.0
		if bits[i] == 1 {
			// 右边 +
			pos += half
		} else {
			// 左边 -
			pos -= half
		}
	}
	return pos, half * 2
}

// 坐标转编码;
// precision 精度，即编码的字符长度
func (geohash GeoHash) GetCode(pnt base.Point2D, precision int) (code string) {
	// 每个字符能代表5个bit，xlen+ylen=precision*5
	// precision为偶数时，xlen=ylen；precision为奇数时，xlen=ylen+1
	ylen := precision * 5 / 2
	xlen := precision*5 - ylen
	xbits := calcOneBits(-180.0, 180.0, pnt.X, int32(xlen), false)
	ybits := calcOneBits(-90.0, 90.0, pnt.Y, int32(ylen), false)
	bits := hashCombineBits(xbits, ybits)
	// bits len一定是5的倍数
	for i := 0; i < len(bits); i += 5 {
		var b32 Base32
		b32.FromBits(bits[i : i+5])
		code += string(b32)
	}
	return
}

// 两个数组交叉合并，注意：xbits可能比ybits多
func hashCombineBits(xbits []byte, ybits []byte) (bits []byte) {
	bits = make([]byte, 0, len(ybits)*2)
	for i, _ := range ybits {
		// 约定：[0 1] 是y方向；[1 0] 是x方向
		bits = append(bits, xbits[i]) // x 放高位
		bits = append(bits, ybits[i]) // y 放低位
	}
	if len(xbits) > len(ybits) {
		bits = append(bits, xbits[len(xbits)-1])
	}
	return
}

// ============================================================= //

type Base32 rune

var DictBase32 = [32]rune{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'k', 'm',
	'n', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

func (base32 Base32) toBits() (bits []byte) {
	value := base32.searchIndex()
	bits = make([]byte, 5)
	for pos := len(bits) - 1; pos >= 0; pos-- {
		bits[pos] = byte(value % 2)
		value /= 2
	}
	return
}

func (base32 Base32) searchIndex() (mid int) {
	// 二分法查找
	low, high := 0, len(DictBase32)-1
	for low <= high {
		mid = (low + high) / 2
		if DictBase32[mid] < rune(base32) {
			low = mid + 1
		} else if DictBase32[mid] > rune(base32) {
			high = mid - 1
		} else {
			return
		}
	}
	return
}

func (this *Base32) FromBits(bits []byte) {
	value := bits2Value(bits)
	*this = Base32(DictBase32[value])
}

// 把二进制形式的数组，转化为十进制数值；如 {1,1,0}-->6
func bits2Value(bits []byte) (value int) {
	bitLen := len(bits)
	for i := 0; i < bitLen; i++ {
		value += int(bits[i]) * base.Power(2, bitLen-i-1)
	}
	return
}
