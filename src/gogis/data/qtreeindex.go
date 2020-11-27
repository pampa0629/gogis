package data

import (
	"fmt"
	"gogis/base"
	"gogis/geometry"
)

// 四叉树索引
type QTreeIndex struct {
	QTreeNode // 根节点
}

// 每个节点控制在多少条记录左右，多了就继续分叉
var ONE_NODE_OBJ_COUNT = 10000

// 四叉树的一个节点
type QTreeNode struct {
	// level int
	ids   []int         // 本节点存的对象id数组
	bboxs []base.Rect2D // 对应的bounds

	bbox    base.Rect2D // 本节点的bounds
	parent  *QTreeNode  // 父节点，根节点的父节点为nil
	isSplit bool        // 是否已分叉，即生成四个子节点
	// 对于叶节点，下面几个为nil
	leftUp    *QTreeNode
	leftDown  *QTreeNode
	rightUp   *QTreeNode
	rightDown *QTreeNode
	// pos       string
}

// 输出
func (this *QTreeNode) String() {
	// fmt.Println("level:", this.level, "pos:", this.pos, "Bbox:", this.bbox, "isSplit:", this.isSplit, "ids'count:", len(this.ids))
}

func (this *QTreeNode) WholeString() {
	// fmt.Println("level:", this.level, "pos:", this.pos, "Bbox:", this.bbox, "isSplit:", this.isSplit, "ids'count:", len(this.ids))
	if this.isSplit {
		this.leftUp.WholeString()
		this.leftDown.WholeString()
		this.rightUp.WholeString()
		this.rightDown.WholeString()
	}
}

func (this *QTreeNode) Init(bbox base.Rect2D, parent *QTreeNode) {
	this.ids = make([]int, 0)
	this.bboxs = make([]base.Rect2D, 0)
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
	this.isSplit = false
	// fmt.Println("QTreeNode.Init()")
	// this.String()
}

// 添加一个对象
func (this *QTreeNode) AddOneGeo(bbox base.Rect2D, id int) {
	// fmt.Println("QTreeNode.AddOneGeo(),bbox:", bbox, "id:", id)
	// this.String()

	if !this.isSplit {
		// 未分叉前，直接加对象即可
		this.AddOneWhenNoSplited(bbox, id)
	} else {
		// 已分叉后，则需要判断geo的bounds来确定是给自己ids，还是往下面的某个子节点中添加
		this.AddOneWhenSplited(bbox, id)
	}
}

// 未分叉时，添加对象
func (this *QTreeNode) AddOneWhenNoSplited(bbox base.Rect2D, id int) {
	// fmt.Println("QTreeNode.AddOneWhenNoSplited(),id:", id)
	// this.String()

	// 未分叉时，先往ids中加
	this.ids = append(this.ids, id)
	this.bboxs = append(this.bboxs, bbox)
	// 直到满了，就分叉
	if len(this.ids) >= ONE_NODE_OBJ_COUNT {
		this.Split()
	}
}

// 创建子节点
func (this *QTreeNode) createChildNodes() {
	// fmt.Println("QTreeNode.createChildNodes()")
	// this.String()

	this.leftUp = new(QTreeNode)
	// this.leftUp.pos = "leftUp"
	this.leftUp.Init(SplitBox(this.bbox, false, true), this)

	this.leftDown = new(QTreeNode)
	// this.leftDown.pos = "leftDown"
	this.leftDown.Init(SplitBox(this.bbox, false, false), this)

	this.rightUp = new(QTreeNode)
	// this.rightUp.pos = "rightUp"
	this.rightUp.Init(SplitBox(this.bbox, true, true), this)

	this.rightDown = new(QTreeNode)
	// this.rightDown.pos = "rightDown"
	this.rightDown.Init(SplitBox(this.bbox, true, false), this)
}

// 分叉；把所管理的所有对象过滤一遍，尽量分配到子节点中
func (this *QTreeNode) Split() {
	// fmt.Println("QTreeNode.Split()")
	// this.String()

	// 先创建子节点
	this.createChildNodes()
	this.isSplit = true

	// 把自己管理的数据，复制一份，并清空自己的
	ids := make([]int, len(this.ids))
	copy(ids, this.ids)
	this.ids = this.ids[0:0]
	bboxs := make([]base.Rect2D, len(this.bboxs))
	copy(bboxs, this.bboxs)
	this.bboxs = this.bboxs[0:0]

	// 再循环处理每个对象
	for i, v := range ids {
		this.AddOneWhenSplited(bboxs[i], v)
	}
}

// 已分叉时，添加对象
func (this *QTreeNode) AddOneWhenSplited(bbox base.Rect2D, id int) {
	// fmt.Println("QTreeNode.AddOneWhenSplited(),,bbox:", bbox, "id:", id)
	// this.String()

	childNode := this.whichChildNode(bbox)
	if childNode != nil {
		// 能放到那个子节点，就尽量下放；可能导致子节点内部发生分叉
		// fmt.Println("find childnode")
		// childNode.String()
		childNode.AddOneGeo(bbox, id)
	} else { // 不能下放，就自己收了
		// fmt.Println("cannot find childnode")
		this.bboxs = append(this.bboxs, bbox)
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
	this.bboxs = this.bboxs[:]
	if this.isSplit {
		nodes := this.getChildNodes()
		for _, v := range nodes {
			v.Clear()
		}
	}
}

// 范围查询，返回id数组
func (this *QTreeNode) Query(bbox base.Rect2D) (ids []int) {
	// 有交集再继续
	if this.bbox.IsIntersect(bbox) {

		ids = make([]int, 0)
		// 查询的时候，一层层处理
		// 先把根节点的都纳入
		ids = append(ids, this.ids...)
		// 有子节点时，每个子节点也都需要处理
		if this.isSplit {
			// center := this.bbox.Center()
			// bboxes := SplitBoxes(bbox, center.X, center.Y)
			nodes := this.getChildNodes()
			for _, v := range nodes {
				ids = append(ids, v.Query(bbox)...)
				// splitNode := this.whichChildNode2(v)
				// if splitNode != nil {
				// 	// splitNode.String() //
				// 	ids = append(ids, splitNode.Query(v)...)
				// } else {
				// 	// 不应该出现的情况
				// 	fmt.Println("error:cannot find child node")
				// 	fmt.Println("i:", i, "box:", v)
				// 	this.String()
				// 	this.leftUp.String()
				// 	this.leftDown.String()
				// 	this.rightUp.String()
				// 	this.rightDown.String()
				// }
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

func (this *QTreeIndex) Init(bbox base.Rect2D, num int) {
	// 暂时考虑四叉树层级不要超过4层，故而叶子节点个数不超过 256个
	// 同时考虑每个叶子节点个数在[1000,10000]范围内
	objCount := int(num / 256)
	ONE_NODE_OBJ_COUNT = base.IntMin(base.IntMax(1000, objCount), 10000)
	this.QTreeNode.Init(bbox, nil)
}

func (this *QTreeIndex) BuildByGeos(geometrys []geometry.Geometry) {
	fmt.Println("QTreeIndex.BuildByGeos")
	for i, geo := range geometrys {
		this.AddOneGeo(geo.GetBounds(), i)
	}
}

func (this *QTreeIndex) BuildByFeas(features []Feature) {
	fmt.Println("QTreeIndex.BuildByFeas")
	for i, fea := range features {
		this.AddOneGeo(fea.Geo.GetBounds(), i)
	}
}

// // 把一个box 按 x/y 切两刀, 可能得到两个小box,也有可能得到四个box；都没切到，返回原bbox
// func SplitBoxes(bbox base.Rect2D, x float64, y float64) (bboxes []base.Rect2D) {
// 	hx, hy := false, false // 是否切中bbox
// 	// 竖着切中了
// 	if bbox.Max.X > x && x > bbox.Min.X {
// 		hx = true
// 	}
// 	// 横着切中了
// 	if bbox.Max.Y > y && y > bbox.Min.Y {
// 		hy = true
// 	}
// 	if hx && hy {
// 		bboxes = append(bboxes, SplitBoxByXY(bbox, x, y)...)
// 	} else if hx {
// 		bboxes = append(bboxes, SplitBoxByX(bbox, x)...)
// 	} else if hy {
// 		bboxes = append(bboxes, SplitBoxByY(bbox, y)...)
// 	} else {
// 		bboxes = append(bboxes, bbox)
// 	}
// 	return
// }

// // 竖着切一刀，分割box，返回左右两个box
// func SplitBoxByX(bbox base.Rect2D, x float64) (bboxes []base.Rect2D) {
// 	bboxes = make([]base.Rect2D, 2)
// 	bboxes[0], bboxes[1] = bbox, bbox
// 	// 0: left
// 	bboxes[0].Max.X = x
// 	// 1: right
// 	bboxes[1].Min.X = x
// 	return
// }

// // 横着切一刀，分割box，返回上下两个box
// func SplitBoxByY(bbox base.Rect2D, y float64) (bboxes []base.Rect2D) {
// 	bboxes = make([]base.Rect2D, 2)
// 	bboxes[0], bboxes[1] = bbox, bbox
// 	// 0: up
// 	bboxes[0].Min.Y = y
// 	// 1: down
// 	bboxes[1].Max.Y = y
// 	return
// }

// // 横着，竖着都能切刀，分割box，返回上下左右四个box
// func SplitBoxByXY(bbox base.Rect2D, x float64, y float64) (bboxes []base.Rect2D) {
// 	bboxes = make([]base.Rect2D, 4)
// 	for i, _ := range bboxes {
// 		bboxes[i] = bbox
// 	}
// 	// 0: leftUp
// 	bboxes[0].Max.X = x
// 	bboxes[0].Min.Y = y
// 	// 1: leftDown
// 	bboxes[1].Max.X = x
// 	bboxes[1].Max.Y = y
// 	// 2: rightUp
// 	bboxes[2].Min.X = x
// 	bboxes[2].Min.Y = y
// 	// 3: rightDown
// 	bboxes[3].Min.X = x
// 	bboxes[3].Max.Y = y
// 	return
// }

// 把一个box从中心点分为四份，返回其中一个bbox, ax/ay 为true时，坐标值增加
func SplitBox(bbox base.Rect2D, ax bool, ay bool) (newBbox base.Rect2D) {
	min := bbox.Min
	center := bbox.Center()
	max := bbox.Max

	if ax && ay {
		newBbox.Min = center
		newBbox.Max = max
	} else if ax {
		newBbox.Min.X = center.X
		newBbox.Max.X = max.X
		newBbox.Min.Y = min.Y
		newBbox.Max.Y = center.Y
	} else if ay {
		newBbox.Min.X = min.X
		newBbox.Max.X = center.X
		newBbox.Min.Y = center.Y
		newBbox.Max.Y = max.Y
	} else {
		newBbox.Min = min
		newBbox.Max = center
	}
	return newBbox
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
