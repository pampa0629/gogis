// Z-Order 索引
package index

import (
	"encoding/binary"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"io"
	"math"
	"sort"
)

// 控制每个最小的cell中，平均对象个数
var ONE_CELL_COUNT = 1000.0

// id和bbox组合在一起
type Idbbox struct {
	Id   int64
	Bbox base.Rect2D
}

type ZOrderIndex struct {
	level    int32 // 划分的层级，0为整体，+1则一分为四
	bbox     base.Rect2D
	code2ids [][]Idbbox // code -->[]id&bbox
}

func (this *ZOrderIndex) Data() [][]Idbbox {
	return this.code2ids
}

func (this *ZOrderIndex) String() {
	fmt.Println("max level:", this.level, "code count:", len(this.code2ids))
}

// 构建后，检查是否有问题；没问题返回true
// id不能重复，即每个id只能存在一个地方
// 每个id的bbox，都必须被code对应的bbox所 cover，且不能被下一级的bbox所cover（否则就应该放到下一级去）
func (this *ZOrderIndex) Check() bool {
	if !this.checkIds() {
		return false
	}
	if !this.checkBbox() {
		return false
	}
	return true
}

// id不能重复存储
func (this *ZOrderIndex) checkIds() bool {
	ids := make(map[int64]byte, len(this.code2ids))
	for i, _ := range this.code2ids {
		for _, vv := range this.code2ids[i] {
			count := len(ids)
			ids[vv.Id] = 0
			if len(ids) == count {
				// 若长度一致，说明id没有写入，存在重复id
				return false
			}
		}
	}
	return true
}

// 每个id的bbox，都必须被code对应的bbox所 cover
// 且不能被下一级的bbox所cover（否则就应该放到下一级去）
func (this *ZOrderIndex) checkBbox() bool {
	for code, idbboxes := range this.code2ids {
		bits := Code2bits(code)
		bbox := this.calcBbox(bits)
		downBitss := buildDownBitss(bits)
		downBboxes := make([]base.Rect2D, 0)
		for _, downBits := range downBitss {
			downBboxes = append(downBboxes, this.calcBbox(downBits))
		}
		for _, v := range idbboxes {
			// 本层级的bbox，必须能cover 所有id的bbox
			if !bbox.IsCover(v.Bbox) {
				return false
			}
			// 如果还有下一层
			if int(this.level) > len(bits)/2 {
				for _, downBbox := range downBboxes {
					// 且id bbox 不能被下一级的bbox所cover
					if downBbox.IsCover(v.Bbox) {
						return false
					}
				}
			}
		}
	}
	return true
}

// 计算得到 level
func CalcZOderLevel(num int64) (level int32) {
	level = int32(math.Log(float64(num)/ONE_CELL_COUNT)/2) + 1
	// 最多15层，int32最多存储16层，还要留一个bit做前置
	level = int32(base.IntMin(int(level), 15))
	level = int32(base.IntMax(int(level), 0)) // 不能小于0
	return
}

// 初始化，之后Add
func (this *ZOrderIndex) Init(bbox base.Rect2D, num int64) {
	this.bbox = bbox
	this.level = CalcZOderLevel(num)
	this.level = 5 // todo

	// 直接用code作为下标，计算所有cell的个数
	cellCount := calcCellCount(this.level)
	this.code2ids = make([][]Idbbox, cellCount)
}

// 初始化，之后直接 QueryDB
func (this *ZOrderIndex) InitDB(bbox base.Rect2D, level int32) {
	this.bbox = bbox
	this.level = level
}

// 输入几何对象，构建索引；下列三种方式等效，同一个对象请勿重复调用Add方法
func (this *ZOrderIndex) AddGeos(geometrys []geometry.Geometry) {
	for _, v := range geometrys {
		this.AddOne(v.GetBounds(), v.GetID())
	}
}

func (this *ZOrderIndex) AddGeo(geo geometry.Geometry) {
	this.AddOne(geo.GetBounds(), geo.GetID())
}

func (this *ZOrderIndex) AddOne(bbox base.Rect2D, id int64) {
	// 计算bbox的code
	code := this.GetCode(bbox)
	// 第一次遇到某个 code，得构建好slice
	if this.code2ids[code] == nil || len(this.code2ids[code]) == 0 {
		this.code2ids[code] = make([]Idbbox, 0)
	}
	this.code2ids[code] = append(this.code2ids[code], Idbbox{id, bbox})
}

// 根据 bits，计算得到 bbox
func (this *ZOrderIndex) calcBbox(bits []byte) (bbox base.Rect2D) {
	bbox.Min = this.bbox.Min
	w := this.bbox.Dx()
	h := this.bbox.Dy()
	for i := 0; i < len(bits); i += 2 {
		// 减半再使用
		w /= 2.0
		h /= 2.0
		if bits[i] == 1 {
			bbox.Min.X += w
		}
		if bits[i+1] == 1 {
			bbox.Min.Y += h
		}
	}
	bbox.Max.X = bbox.Min.X + w
	bbox.Max.Y = bbox.Min.Y + h

	return
}

func (this *ZOrderIndex) GetCode(bbox base.Rect2D) int32 {
	bits := this.calcBboxBits(bbox)
	return Bits2code(bits)
}

func (this *ZOrderIndex) calcBboxBits(bbox base.Rect2D) (bits []byte) {
	bits = make([]byte, 0)
	// 思路：先分别计算 min和max两个点的code，按双数对比，提取前2*N位都一样的部分
	minBits := this.calcPointBits(bbox.Min, true)
	maxBits := this.calcPointBits(bbox.Max, false)
	// fmt.Println("minBits:", minBits, "maxBits:", maxBits)
	for i := 0; i < len(minBits); i += 2 {
		if minBits[i] == maxBits[i] && minBits[i+1] == maxBits[i+1] {
			bits = append(bits, minBits[i:i+2]...)
		} else { // 一旦不一样，就返回，后面的bits相等没有作用
			break
		}
	}
	return
}

// bits转为code
// 按照层级从高到低排列，同层级按照Z order排列
// 序号从0起，0即为level 0的code
func Bits2code(bits []byte) (code int32) {
	level := int32(len(bits) / 2)
	code = calcCellCount(level - 1) // 到上一层为止的cell个数
	bitsLen := len(bits)
	for i := 0; i < bitsLen; i += 2 {
		code += int32(bits[i]) << (bitsLen - i - 1)
		code += int32(bits[i+1]) << (bitsLen - i - 2)
	}
	return
}

// 计算得到 level 和剩余的code
func getLevelAndRemain(code int) (int, int) {
	// 先算出是第几层的
	level := 0
	for {
		count := int(math.Pow(4.0, float64(level)))
		if code < count {
			break
		}
		code -= count // 剩余的code
		level++
	}
	return level, code
}

// code转为bits
func Code2bits(code int) (bits []byte) {
	level, remain := getLevelAndRemain(code)
	length := level * 2
	bits = make([]byte, length)
	hit := 1
	for i := 0; i < length; i++ {
		if remain&hit > 0 {
			bits[length-i-1] = 1
		}
		hit = hit << 1 // 往前做位移
	}
	return
}

// 计算一个点在最小cell（最大level）中的bits
func (this *ZOrderIndex) calcPointBits(pnt base.Point2D, isMin bool) (bits []byte) {
	xbits := this.calcOneBits(this.bbox.Min.X, this.bbox.Dx()/2, pnt.X, this.level, isMin)
	ybits := this.calcOneBits(this.bbox.Min.Y, this.bbox.Dy()/2, pnt.Y, this.level, isMin)
	bits = this.combineBits(xbits, ybits)
	return
}

// 两个数组交叉合并
func (this *ZOrderIndex) combineBits(xbits []byte, ybits []byte) (bits []byte) {
	bits = make([]byte, 0, len(xbits)*2)
	for i, _ := range xbits {
		// 约定：0 1 是y方向；1 0是x方向
		bits = append(bits, xbits[i]) // x 放高位
		bits = append(bits, ybits[i]) // y 放低位
	}
	return
}

// 计算一个方向的编码
func (this *ZOrderIndex) calcOneBits(zero, halfLength, pos float64, level int32, isMin bool) (bits []byte) {
	// 小为0，大为1
	for level > 0 {
		// isMin 为true，意味着输入的是 bbox的min，大于等于 就能 为1
		// isMin 为false，意味着输入的是 bbox的max，必须 大于 才能 为1
		// 即：bbox的min，要放松要求，或者等于给ta加一个极小值；而对于max，则必须严格要求
		if (isMin && base.IsBigEqual(pos, zero+halfLength)) || (!isMin && (pos > zero+halfLength)) {
			bits = append(bits, 1)
			zero += halfLength
		} else {
			bits = append(bits, 0)
		}
		level--
		halfLength /= 2
	}
	return
}

// 查询得到 bbox 所涉及的code，返回code数组（已排序）
func (this *ZOrderIndex) QueryDB(bbox base.Rect2D) (codes []int32) {
	// 思路：
	// 先找高层的level的code，注意需要做box判断
	// 再迭代处理本层和低层
	//   得到bbox的code，
	//   低层的level，则先判断是否bbox相交，再迭代查询，直到最底层的level为止

	bits := this.calcBboxBits(bbox)

	// 更高层的level
	upbits := make([]byte, len(bits))
	copy(upbits, bits)
	for len(upbits) > 0 {
		upbits = upbits[0 : len(upbits)-2] // 去掉最后两个bit，即提升一个level
		upcode := Bits2code(upbits)
		codes = append(codes, upcode)
	}

	// 这里查询本层和下层的
	codes = append(codes, this.queryThisDownDB(bbox, bits)...)

	sort.Sort(base.Int32s(codes))
	return
}

// 查询本层以及下层的，需要迭代执行，直到最底层
func (this *ZOrderIndex) queryThisDownDB(bbox base.Rect2D, bits []byte) (codes []int32) {
	// 得到本层的 code
	code := Bits2code(bits)
	codes = append(codes, code)

	// fmt.Println("level:", len(bits)/2, "bits:", bits, "ids:", ids, "code:", code, "cell bbox:", cellBbox, "bbox:", bbox)

	// 若存在更低层的level，则构造下层的bits，判断bbox是否相交，再做查询
	if len(bits)/2 < int(this.level) { // 不是最底层
		downBitss := buildDownBitss(bits)
		// fmt.Println("bits:", bits, "downBitss:", downBitss)
		for _, downBits := range downBitss {
			downBbox := this.calcBbox(downBits)
			if bbox.IsIntersect(downBbox) {
				downCode := this.queryThisDownDB(bbox, downBits)
				codes = append(codes, downCode...)
			}
		}
	}

	return
}

// 范围查询，返回id数组
func (this *ZOrderIndex) Query(bbox base.Rect2D) (ids []int64) {
	// 思路：
	// 先找高层的level的code，注意需要做box判断
	// 再迭代处理本层和低层
	//   得到bbox的code，这个code的ids，再进一步做 box判断
	//   低层的level，则先判断是否bbox相交，再迭代查询，直到最底层的level为止

	bits := this.calcBboxBits(bbox)

	// 更高层的level
	upbits := make([]byte, len(bits))
	copy(upbits, bits)
	for len(upbits) > 0 {
		upbits = upbits[0 : len(upbits)-2] // 去掉最后两个bit，即提升一个level
		upcode := Bits2code(upbits)
		upBbox := this.calcBbox(upbits)
		ids = append(ids, this.getIdsByBbox(bbox, upBbox, upcode)...)
	}

	// 这里查询本层和下层的
	ids = append(ids, this.queryThisDown(bbox, bits)...)

	return
}

// 查询本层以及下层的，需要迭代执行，直到最底层
func (this *ZOrderIndex) queryThisDown(bbox base.Rect2D, bits []byte) (ids []int64) {
	// 得到本层的 code
	code := Bits2code(bits)
	// 计算本cell的bbox
	cellBbox := this.calcBbox(bits)
	// 同层级的一个cell，即code相等
	ids = this.getIdsByBbox(bbox, cellBbox, code)

	// 若存在更低层的level，则构造下层的bits，判断bbox是否相交，再做查询
	if len(bits)/2 < int(this.level) { // 不是最底层
		downBitss := buildDownBitss(bits)
		for _, downBits := range downBitss {
			downBbox := this.calcBbox(downBits)
			if bbox.IsIntersect(downBbox) {
				downIds := this.queryThisDown(bbox, downBits)
				ids = append(ids, downIds...)
			}
		}
	}

	return
}

// 构造下层的四个bits
func buildDownBitss(bits []byte) (downBitss [][]byte) {
	bitCount := len(bits)
	downBitss = make([][]byte, 4)
	for i, _ := range downBitss {
		downBitss[i] = make([]byte, bitCount+2)
		// 前面的bit一模一样
		copy(downBitss[i][0:bitCount], bits[:])
		// 后面两位按照 00 01 10 11 来构造
		downBitss[i][bitCount] = byte(i / 2)
		downBitss[i][bitCount+1] = byte(i % 2)
	}
	return
}

// 根据code，加上box判断，添加ids
func (this *ZOrderIndex) getIdsByBbox(bbox, cellBbox base.Rect2D, code int32) (ids []int64) {
	// box覆盖，就全部拿下
	if bbox.IsCover(cellBbox) {
		for _, v := range this.code2ids[code] {
			ids = append(ids, v.Id)
		}
	} else {
		// 否则就还得 一个个box的判断
		for _, v := range this.code2ids[code] {
			if bbox.IsIntersect(v.Bbox) {
				ids = append(ids, v.Id)
			}
		}
	}
	return
}

// 清空
func (this *ZOrderIndex) Clear() {
	this.code2ids = this.code2ids[:0]
}

// 返回自己的类型
func (this *ZOrderIndex) Type() SpatialIndexType {
	return TypeZOrderIndex
}

// 保存和加载，避免每次都要重复构建
func (this *ZOrderIndex) Save(w io.Writer) {
	binary.Write(w, binary.LittleEndian, this.level)
	binary.Write(w, binary.LittleEndian, this.bbox)

	cellCount := int32(len(this.code2ids))
	binary.Write(w, binary.LittleEndian, cellCount)
	for _, v := range this.code2ids {
		idCount := int32(len(v))
		binary.Write(w, binary.LittleEndian, idCount)
		binary.Write(w, binary.LittleEndian, v)
	}
}

func (this *ZOrderIndex) Load(r io.Reader) {
	binary.Read(r, binary.LittleEndian, &this.level)
	binary.Read(r, binary.LittleEndian, &this.bbox)

	var cellCount int32
	binary.Read(r, binary.LittleEndian, &cellCount)

	this.code2ids = make([][]Idbbox, cellCount)
	for i, _ := range this.code2ids {
		var idCount int32
		binary.Read(r, binary.LittleEndian, &idCount)
		this.code2ids[i] = make([]Idbbox, idCount)
		binary.Read(r, binary.LittleEndian, this.code2ids[i])
	}
	this.String()
}
