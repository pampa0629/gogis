package index

import (
	"encoding/binary"
	"gogis/base"
	"gogis/geometry"
	"io"
)

// 每个节点控制在多少条记录左右，多了就继续分叉
var ONE_NODE_OBJ_COUNT = 10000

// 四叉树索引
type QTreeIndex struct {
	QTreeNode // 根节点
}

func (this *QTreeIndex) Type() SpatialIndexType {
	return TypeQTreeIndex
}

func (this *QTreeIndex) Init(bbox base.Rect2D, num int64) {
	// 暂时考虑四叉树层级不要超过4层，故而叶子节点个数不超过 256个
	// 同时考虑每个叶子节点个数在[1000,10000]范围内
	objCount := int(num / 256)
	ONE_NODE_OBJ_COUNT = base.IntMin(base.IntMax(1000, objCount), 10000)
	this.QTreeNode.Init(bbox, nil)
}

func (this *QTreeIndex) AddGeos(geometrys []geometry.Geometry) {
	// fmt.Println("QTreeIndex.BuildByGeos")
	for _, geo := range geometrys {
		this.AddOne(geo.GetBounds(), geo.GetID())
	}
}

func (this *QTreeIndex) AddGeo(geo geometry.Geometry) {
	this.AddOne(geo.GetBounds(), geo.GetID())
}

// func (this *QTreeIndex) BuildByFeas(features []data.Feature) {
// 	fmt.Println("QTreeIndex.BuildByFeas")
// 	for i, fea := range features {
// 		this.AddOneGeo(fea.Geo.GetBounds(), int64(i))
// 	}
// }

// 保存
func (this *QTreeNode) Save(w io.Writer) {
	binary.Write(w, binary.LittleEndian, base.Bool2Int32(this.isSplited))
	binary.Write(w, binary.LittleEndian, this.bbox)
	count := int32(len(this.ids))
	binary.Write(w, binary.LittleEndian, count)
	binary.Write(w, binary.LittleEndian, this.ids)
	binary.Write(w, binary.LittleEndian, this.bboxes)
	if this.isSplited {
		nodes := this.getChildNodes()
		for _, v := range nodes {
			v.Save(w)
		}
	}
}

// 加载
func (this *QTreeNode) Load(r io.Reader) {
	var isSplited int32
	binary.Read(r, binary.LittleEndian, &isSplited)
	this.isSplited = isSplited != 0
	binary.Read(r, binary.LittleEndian, &this.bbox)
	var count int32
	binary.Read(r, binary.LittleEndian, &count)
	this.ids = make([]int64, count)
	this.bboxes = make([]base.Rect2D, count)
	binary.Read(r, binary.LittleEndian, this.ids)
	binary.Read(r, binary.LittleEndian, this.bboxes)
	if this.isSplited {
		this.createChildNodes()
		nodes := this.getChildNodes()
		for _, v := range nodes {
			v.Load(r)
		}
	}
}

// 四叉树的一个节点
type QTreeNode struct {
	// level int
	ids    []int64       // 本节点存的对象id数组
	bboxes []base.Rect2D // 对应的bounds

	bbox      base.Rect2D // 本节点的bounds
	parent    *QTreeNode  // 父节点，根节点的父节点为nil
	isSplited bool        // 是否已分叉，即生成四个子节点
	// 对于叶节点，下面几个为nil
	leftUp    *QTreeNode
	leftDown  *QTreeNode
	rightUp   *QTreeNode
	rightDown *QTreeNode
	// pos       string
}

// 构建后，检查是否有问题；没问题返回true
// 可能存在的问题：
// 是否分叉，与节点指针是否为空
// 子节点的父节点，是否等于自己
// 判断bbox的范围
func (this *QTreeNode) Check() bool {
	// 检查分叉与节点指针的关系
	if this.isSplited {
		if this.leftUp == nil || this.leftDown == nil || this.rightUp == nil || this.rightDown == nil {
			return false
		}
	} else {
		if this.leftUp != nil || this.leftDown != nil || this.rightUp != nil || this.rightDown != nil {
			return false
		}
	}

	// 检查子节点的父节点，是否等于自己
	if this.isSplited {
		nodes := this.getChildNodes()
		for _, v := range nodes {
			if v.parent != this {
				return false
			}
		}
	}

	// bboxes的并，要被bbox所覆盖
	if len(this.bboxes) > 0 {
		if !this.bbox.IsCover(base.UnionBounds(this.bboxes)) {
			return false
		}
	}

	// 子节点bbox的并，应等于 bbox
	if this.isSplited {
		var bbox base.Rect2D
		bbox.Init()
		nodes := this.getChildNodes()
		for _, v := range nodes {
			bbox.Union(v.bbox)
		}
		if bbox != this.bbox {
			return false
		}
	}

	// ids和bboxes的长度要相等
	if len(this.ids) != len(this.bboxes) {
		return false
	}

	// 若分叉，子节点也要做同样的检查
	if this.isSplited {
		nodes := this.getChildNodes()
		for _, v := range nodes {
			if !v.Check() {
				return false
			}
		}
	}

	return true
}

// 输出
func (this *QTreeNode) string() {
	// fmt.Println("level:", this.level, "pos:", this.pos, "Bbox:", this.bbox, "isSplited:", this.isSplited, "ids'count:", len(this.ids))
}

func (this *QTreeNode) wholeString() {
	this.string()
	// fmt.Println("level:", this.level, "pos:", this.pos, "Bbox:", this.bbox, "isSplited:", this.isSplited, "ids'count:", len(this.ids))
	if this.isSplited {
		this.leftUp.wholeString()
		this.leftDown.wholeString()
		this.rightUp.wholeString()
		this.rightDown.wholeString()
	}
}

func (this *QTreeNode) Init(bbox base.Rect2D, parent *QTreeNode) {
	this.ids = make([]int64, 0)
	this.bboxes = make([]base.Rect2D, 0)
	this.parent = parent
	// if parent == nil {
	// 	this.level = 0
	// 	this.pos = "root"
	// } else {
	// 	this.level = parent.level + 1
	// }
	this.bbox = bbox
	this.leftUp = nil
	this.leftDown = nil
	this.rightDown = nil
	this.rightUp = nil
	this.isSplited = false
	// fmt.Println("QTreeNode.Init()")
	// this.String()
}

// 添加一个对象
func (this *QTreeNode) AddOne(bbox base.Rect2D, id int64) {
	// fmt.Println("QTreeNode.addOne(),bbox:", bbox, "id:", id)
	// this.String()

	if !this.isSplited {
		// 未分叉前，直接加对象即可
		this.addOneWhenNoSplited(bbox, id)
	} else {
		// 已分叉后，则需要判断geo的bounds来确定是给自己ids，还是往下面的某个子节点中添加
		this.addOneWhenSplited(bbox, id)
	}
}

// 未分叉时，添加对象
func (this *QTreeNode) addOneWhenNoSplited(bbox base.Rect2D, id int64) {
	// fmt.Println("QTreeNode.addOneWhenNoSplited(),id:", id)
	// this.String()

	// 未分叉时，先往ids中加
	this.ids = append(this.ids, id)
	this.bboxes = append(this.bboxes, bbox)
	// 直到满了，就分叉
	if len(this.ids) >= ONE_NODE_OBJ_COUNT {
		this.split()
	}
}

// 创建子节点
func (this *QTreeNode) createChildNodes() {
	// fmt.Println("QTreeNode.createChildNodes()")
	// this.String()

	this.leftUp = new(QTreeNode)
	// this.leftUp.pos = "leftUp"
	this.leftUp.Init(base.SplitBounds(this.bbox, false, true), this)

	this.leftDown = new(QTreeNode)
	// this.leftDown.pos = "leftDown"
	this.leftDown.Init(base.SplitBounds(this.bbox, false, false), this)

	this.rightUp = new(QTreeNode)
	// this.rightUp.pos = "rightUp"
	this.rightUp.Init(base.SplitBounds(this.bbox, true, true), this)

	this.rightDown = new(QTreeNode)
	// this.rightDown.pos = "rightDown"
	this.rightDown.Init(base.SplitBounds(this.bbox, true, false), this)
}

// 分叉；把所管理的所有对象过滤一遍，尽量分配到子节点中
func (this *QTreeNode) split() {
	// fmt.Println("QTreeNode.split()")
	// this.String()

	// 先创建子节点
	this.createChildNodes()
	this.isSplited = true

	// 把自己管理的数据，复制一份，并清空自己的
	ids := make([]int64, len(this.ids))
	copy(ids, this.ids)
	this.ids = this.ids[0:0]
	bboxes := make([]base.Rect2D, len(this.bboxes))
	copy(bboxes, this.bboxes)
	this.bboxes = this.bboxes[0:0]

	// 再循环处理每个对象
	for i, v := range ids {
		this.addOneWhenSplited(bboxes[i], v)
	}
}

// 已分叉时，添加对象
func (this *QTreeNode) addOneWhenSplited(bbox base.Rect2D, id int64) {
	// fmt.Println("QTreeNode.addOneWhenSplited(),,bbox:", bbox, "id:", id)
	// this.String()

	childNode := this.whichChildNode(bbox)
	if childNode != nil {
		// 能放到那个子节点，就尽量下放；可能导致子节点内部发生分叉
		// fmt.Println("find childnode")
		// childNode.String()
		childNode.AddOne(bbox, id)
	} else { // 不能下放，就自己收了
		// fmt.Println("cannot find childnode")
		this.bboxes = append(this.bboxes, bbox)
		this.ids = append(this.ids, id)
	}
}

// 判断某个对象应该放到哪个子节点中，或者不能下放(返回nil)
// 当只被单个子节点覆盖时，才能放给TA
func (this *QTreeNode) whichChildNode(bbox base.Rect2D) *QTreeNode {
	// fmt.Println("QTreeNode.whichChildNode()")

	nodes := this.getChildNodes()
	for _, v := range nodes {
		// 允许两个矩形的边界相交
		if v.bbox.IsCover(bbox) {
			return v
		}
	}
	return nil
}

func (this *QTreeNode) getChildNodes() (nodes []*QTreeNode) {
	// fmt.Println("QTreeNode.getChildNodes()")
	nodes = make([]*QTreeNode, 4)
	nodes[0] = this.leftUp
	nodes[1] = this.leftDown
	nodes[2] = this.rightUp
	nodes[3] = this.rightDown
	return
}

// 清空；同时迭代清空所管理的子节点
func (this *QTreeNode) Clear() {
	this.ids = this.ids[:]
	this.bboxes = this.bboxes[:]
	if this.isSplited {
		nodes := this.getChildNodes()
		for _, v := range nodes {
			v.Clear()
		}
	}
}

// 范围查询，返回id数组
func (this *QTreeNode) Query(bbox base.Rect2D) (ids []int64) {
	// 有交集再继续
	if this.bbox.IsIntersect(bbox) {

		// ids = make([]int64, 0)
		// 查询的时候，一层层处理
		// 先把根节点的都纳入
		// ids = append(ids, this.ids...) // 模糊查找
		// 精确查找
		for i, v := range this.ids {
			if this.bboxes[i].IsIntersect(bbox) {
				ids = append(ids, v)
			}
		}

		// 有子节点时，每个子节点也都需要处理
		if this.isSplited {
			nodes := this.getChildNodes()
			for _, v := range nodes {
				ids = append(ids, v.Query(bbox)...)
			}
		}
	}

	return
}

// 查询时，判断采用哪个子节点
// 区别在于：构建索引时，对象必须全部落到 节点的bounds中；
// 而查询范围,则是两个矩形有交叠（有范围相交）即OK
func (this *QTreeNode) whichChildNode2(bbox base.Rect2D) *QTreeNode {
	// fmt.Println("QTreeNode.whichChildNode2()")

	nodes := this.getChildNodes()
	for _, v := range nodes {
		if v.bbox.IsOverlap(bbox) {
			return v
		}
	}
	return nil
}

// // 把一个box从中心点分为四份，返回其中一个bbox, ax/ay 为true时，坐标值增加
// func SplitBox2(bbox base.Rect2D, ax bool, ay bool) (newBbox base.Rect2D) {
// 	// 先算出左下角的bbox
// 	newBbox.Min = bbox.Min
// 	newBbox.Max = bbox.Center()
// 	// 还算出bounds的长度和宽度
// 	dx := newBbox.Dx()
// 	dy := newBbox.Dy()
// 	if ax {
// 		newBbox.Min.X += dx
// 		newBbox.Max.X += dx
// 	}
// 	if ay {
// 		newBbox.Min.Y += dy
// 		newBbox.Max.Y += dy
// 	}
// 	return newBbox
// }
