package data

import (
	"fmt"
	"gogis/base"
	"gogis/geometry"
)

// 每个节点控制在多少个子节点(子对象)范围
// var RTREE_MIN_OBJ_COUNT = 1000
var RTREE_MAX_OBJ_COUNT = 10000

// R树索引
type RTreeIndex struct {
	RTreeNode // 根节点
}

func (this *RTreeIndex) Init(bbox base.Rect2D, num int) {
	this.RTreeNode.Init(bbox)
	objCount := num / 16 // 为啥除以这个数，我也是不知道
	RTREE_MAX_OBJ_COUNT = base.IntMax(base.IntMin(objCount, 100000), 10000)
	fmt.Println("rtree node obj count:", RTREE_MAX_OBJ_COUNT)
}

func (this *RTreeIndex) BuildByGeos(geometrys []geometry.Geometry) {
	for i, geo := range geometrys {
		this.AddOneGeo(geo.GetBounds(), i)
	}
	// this.WholeString()
}

func (this *RTreeIndex) BuildByFeas(features []Feature) {
	for i, fea := range features {
		this.AddOneGeo(fea.Geo.GetBounds(), i)
	}
	// this.WholeString()
}

// 范围查询，返回id数组
func (this *RTreeIndex) Query(bbox base.Rect2D) []int {
	return this.RTreeNode.Query(bbox)
}

func (this *RTreeIndex) Clear() {
	this.RTreeNode.Clear()
}

// R树的一个节点
type RTreeNode struct {
	level  int         // 层级
	parent *RTreeNode  // 父节点，根节点的父节点为nil
	bbox   base.Rect2D // 本节点的bounds
	isLeaf bool        // 是否为叶子节点；如果是叶子节点，则没有子节点，存储对象（id和bbox）数组；非子节点时，存储子节点数组

	nodes []*RTreeNode // 子节点

	ids    []int         // 对象id
	bboxes []base.Rect2D // 对象bounds
}

func (this *RTreeNode) String() {
	fmt.Println("level:", this.level, ", isLeaf:", this.isLeaf, ", bbox:", this.bbox, ", nodes'count:", len(this.nodes), ", ids's count:", len(this.ids))
	// fmt.Println("")
}
func (this *RTreeNode) WholeString() {
	this.String()
	for _, v := range this.nodes {
		v.String()
	}
	// fmt.Println("isRoot:", this.parent == nil, "isLeaf:", this.isLeaf, "nodes'count:", len(this.nodes), "bboxes's count:", len(this.bboxes), "ids's count:", len(this.ids))
	// fmt.Println("")
}

func (this *RTreeNode) Init(bbox base.Rect2D) {
	this.level = 0
	this.parent = nil
	this.bbox = bbox
	this.isLeaf = true
}

func (this *RTreeNode) AddOneGeo(bbox base.Rect2D, id int) {
	// fmt.Println("RTreeNode.AddOneGeo()", bbox, id)
	// this.String()

	// 过程：先看自己是否叶子节点；
	//     若是，则给自己添加对象，添加后，若超过最大个数，则引发自己分裂为两个叶子节点；
	//         怎么个分裂法，得看分裂后，两个节点的范围尽可能的分开，且面积总和最小
	//         之后还得看父节点所管理的子节点个数是否超过最大个数，若是，则引发父节点分裂；往上以此类推
	//     若不是叶子节点，则往下找一个最合适的节点，以此类推，直到找到最合适的叶子节点来存放（之后有可能引发自己的分裂）
	//         最合适的标准为：加入后，使得节点范围的面积增加值最小
	if this.isLeaf {
		this.ids = append(this.ids, id)
		this.bboxes = append(this.bboxes, bbox)
		this.bbox.Union(bbox)
		if len(this.ids) >= RTREE_MAX_OBJ_COUNT {
			this.Split()
		}
	} else {
		node := this.findBestChild(bbox)
		node.AddOneGeo(bbox, id)
	}
}

// 找到最合适的子节点，以便往该子节点中增加对象
// 标准为：对象加入后，使得节点范围的面积增加值最小
func (this *RTreeNode) findBestChild(bbox base.Rect2D) *RTreeNode {
	// fmt.Println("RTreeNode.findBestChild()", bbox)
	// this.String()

	maxMoreArea := -1.0 // 新增加的面积
	maxNode := (*RTreeNode)(nil)
	for _, node := range this.nodes {
		newBbox := node.bbox
		newBbox.Union(bbox)
		moreArea := newBbox.Area() - node.bbox.Area()
		if moreArea > maxMoreArea {
			maxMoreArea = moreArea
			maxNode = node
		}
	}
	return maxNode
}

// 节点分裂为两个同级别的节点
// 标准为：两个节点的范围应尽可能分离，且面积总和最小
// 分裂后，若导致父节点所管理的子节点数量超过最大值，父节点也得进一步分裂；迭代往复
// 注意：本节点可能为叶子节点，也可能是中间节点，还有可能是根节点，根节点也有可能同时就是叶子节点；得分别处理
func (this *RTreeNode) Split() {
	// fmt.Println("RTreeNode.Split()")
	// this.String()

	// 若自己就是根节点，则自己不能动，把从自己分裂出来的两个节点纳入自己管辖
	leftNode := (*RTreeNode)(nil)
	rightNode := (*RTreeNode)(nil)

	if this.isLeaf {
		// 叶子节点分裂，把自己变为两个，划分好对象
		leftNode, rightNode = this.SplitLeaf()
	} else {
		// 中间节点分裂，是把自己的子节点分成两拨，分别归到新生成的两个同级节点中
		leftNode, rightNode = this.SplitNoLeaf()
	}

	if this.parent == nil {
		// 若自己是根节点，先清空自己（注意不能调用Clear方法）
		// 再给自己添加新生成的节点
		this.ids = this.ids[0:0]
		this.bboxes = this.bboxes[0:0]
		this.nodes = this.nodes[0:0]
		this.isLeaf = false
		leftNode.parent = this
		rightNode.parent = this
		leftNode.level = 1
		rightNode.level = 1
		this.nodes = append(this.nodes, leftNode)
		this.nodes = append(this.nodes, rightNode)
	} else {
		// 若自己非根节点，从父节点中去掉自己，添加两个新节点（同级别）
		// 1,从父节点中移除自己
		// a = append(a[:i], a[i+1:]...) // 删除中间1个元素
		for i, v := range this.parent.nodes {
			if v == this {
				this.parent.nodes = append(this.parent.nodes[:i], this.parent.nodes[i+1:]...)
				break
			}
		}

		// 2, 给父节点添加两个新生成的节点
		leftNode.parent = this.parent
		rightNode.parent = this.parent
		leftNode.level = this.level
		rightNode.level = this.level
		this.parent.nodes = append(this.parent.nodes, leftNode)
		this.parent.nodes = append(this.parent.nodes, rightNode)
		// 3，再判断是否会引发父节点的分裂
		if len(this.parent.nodes) >= RTREE_MAX_OBJ_COUNT {
			this.parent.Split()
		}
	}
}

// 分裂非叶子节点
func (this *RTreeNode) SplitNoLeaf() (left, right *RTreeNode) {
	// fmt.Println("RTreeNode.SplitNoLeaf()")
	// this.String()

	leftNodes, rightNodes, leftBbox, rightBbox := SplitNodes(this.nodes)
	left = new(RTreeNode)
	left.isLeaf = false
	left.nodes = leftNodes
	left.bbox = leftBbox

	right = new(RTreeNode)
	right.isLeaf = false
	right.nodes = rightNodes
	right.bbox = rightBbox
	return
}

// 把节点数组分为两个
func SplitNodes(nodes []*RTreeNode) (leftNodes, rightNodes []*RTreeNode, leftBbox, rightBbox base.Rect2D) {
	// fmt.Println("RTreeNode.SplitNodes(), nodes's count:", len(nodes))
	// this.String()

	// 方法是：先找到两个距离最远的bbox，剩余的bbox分别根据合并后面积增量最小原则选择
	leftNode, rightNode := findSeedNode(nodes)
	leftBbox = leftNode.bbox
	rightBbox = rightNode.bbox
	leftNodes = append(leftNodes, leftNode)
	rightNodes = append(rightNodes, rightNode)
	for _, v := range nodes {
		// 防止重复加入
		if v != leftNode && v != rightNode {
			leftMore := calcMoreArea(leftBbox, v.bbox)
			rightMore := calcMoreArea(rightBbox, v.bbox)
			// 谁导致增加量小，就给谁
			if leftMore < rightMore {
				leftNodes = append(leftNodes, v)
				leftBbox.Union(v.bbox)
			} else {
				rightNodes = append(rightNodes, v)
				rightBbox.Union(v.bbox)
			}
		}
	}
	return
}

// 把对象数组分为两个
func SplitBoxes(bboxes []base.Rect2D, ids []int) (leftBboxes []base.Rect2D, leftIds []int, rightBboxes []base.Rect2D, rightIds []int) {
	// fmt.Println("RTreeNode.SplitBoxes(), boxes's count:", len(bboxes), ",ids's count:", len(ids))
	// this.String()

	// 方法是：先找到两个距离最远的bbox，剩余的bbox分别根据合并后面积增量最小原则选择
	leftBox, leftId, rightBox, rightId := findSeedBoxes(bboxes, ids)
	// fmt.Println("boxes:", bboxes)
	// fmt.Println("ids:", ids)
	// fmt.Println("res left:", leftBox, leftId)
	// fmt.Println("res right:", rightBox, rightId)

	leftBboxes = append(leftBboxes, leftBox)
	leftIds = append(leftIds, leftId)
	rightBboxes = append(rightBboxes, rightBox)
	rightIds = append(rightIds, rightId)
	for i, v := range bboxes {
		// 防止重复加入
		if ids[i] != leftId && ids[i] != rightId {
			leftMore := calcMoreArea(leftBox, v)
			rightMore := calcMoreArea(rightBox, v)
			// 谁导致增加量小，就给谁
			if leftMore < rightMore {
				leftBboxes = append(leftBboxes, v)
				leftIds = append(leftIds, ids[i])
			} else {
				rightBboxes = append(rightBboxes, v)
				rightIds = append(rightIds, ids[i])
			}
		}
	}
	return
}

// 计算合并后面积的增量
func calcMoreArea(bbox1, bbox2 base.Rect2D) float64 {
	newBbox := bbox1.Clone()
	newBbox.Union(bbox2)
	return newBbox.Area() - bbox2.Area()
}

// 找到两个种子，要求是两个距离足够远
func findSeedBoxes(bboxes []base.Rect2D, ids []int) (leftBox base.Rect2D, leftId int, rightBox base.Rect2D, rightId int) {
	// fmt.Println("RTreeNode.findSeedBoxes(), boxes's count:", len(bboxes), ",ids's count:", len(ids))

	count := len(ids)
	// maxDist := make([]float64, count)
	// maxId := make([]int, count)
	// maxBox := make([]base.Rect2D, count)
	// 先得到所有bounds的中心点
	centerPnts := make([]base.Point2D, count)
	for i, v := range bboxes {
		centerPnts[i] = v.Center()
	}

	// 合并所有bbox
	wholeBbox := UnionBboxes(bboxes)
	leftPos := findNeareastPoint(wholeBbox.Min, centerPnts)
	rightPos := findNeareastPoint(wholeBbox.Max, centerPnts)
	if leftPos == rightPos {
		fmt.Println("出现特例了")
		if leftPos > 0 {
			rightPos = 0
		} else {
			rightPos = 1
		}
	}

	leftId = ids[leftPos]
	rightId = ids[rightPos]
	leftBox = bboxes[leftPos]
	rightBox = bboxes[rightPos]

	return
}

// 找到数组中，距离最近的点
func findNeareastPoint(pnt base.Point2D, pnts []base.Point2D) (pos int) {
	pos = 0
	minDist := pnt.DistanceSquare(pnts[0])
	for i := 1; i < len(pnts); i++ {
		dist := pnt.DistanceSquare(pnts[i])
		if dist < minDist {
			minDist = dist
			pos = i
		}
	}
	return pos
}

// 找到两个种子，要求是两个距离足够远
func findSeedBoxes2(bboxes []base.Rect2D, ids []int) (leftBox base.Rect2D, leftId int, rightBox base.Rect2D, rightId int) {
	// fmt.Println("RTreeNode.findSeedBoxes2(), boxes's count:", len(bboxes), ",ids's count:", len(ids))

	count := len(ids)
	maxDist := make([]float64, count)
	maxId := make([]int, count)
	maxBox := make([]base.Rect2D, count)
	// 先得到所有bounds的中心点
	centerPnts := make([]base.Point2D, count)
	for i, v := range bboxes {
		centerPnts[i] = v.Center()
	}

	// 以此计算距离，并存储起来
	for i := 0; i < count; i++ {
		for j := i + 1; j < count; j++ {
			dist := centerPnts[i].DistanceSquare(centerPnts[j])
			if dist > maxDist[i] {
				maxDist[i] = dist
				maxId[i] = ids[j]
				maxBox[i] = bboxes[j]
			}
		}
	}

	// 得到最大值的序号
	no := base.Max(maxDist) // no --> i

	leftId = ids[no]
	rightId = maxId[no]
	leftBox = bboxes[no]
	rightBox = maxBox[no]

	return
}

// 找到两个种子节点，要求是两个距离足够远
func findSeedNode(nodes []*RTreeNode) (leftNode, rightNode *RTreeNode) {
	// fmt.Println("RTreeNode.findSeedNode()")

	bboxes := make([]base.Rect2D, len(nodes))
	ids := make([]int, len(nodes))
	for i, v := range nodes {
		bboxes[i] = v.bbox
		ids[i] = i
	}
	_, leftId, _, rightId := findSeedBoxes(bboxes, ids)
	leftNode = nodes[leftId]
	rightNode = nodes[rightId]
	return
}

// 分裂叶子节点
func (this *RTreeNode) SplitLeaf() (left, right *RTreeNode) {
	// fmt.Println("RTreeNode.SplitLeaf()")
	// this.String()

	leftBboxes, leftIds, rightBboxes, rightIds := SplitBoxes(this.bboxes, this.ids)

	left = new(RTreeNode)
	left.isLeaf = true
	left.bboxes = leftBboxes
	left.ids = leftIds
	left.bbox = UnionBboxes(leftBboxes) // 合并bounds
	// fmt.Println("left bboxes'count:", len(left.bboxes), " ids'count:", len(left.ids))

	right = new(RTreeNode)
	right.isLeaf = true
	right.bboxes = rightBboxes
	right.ids = rightIds
	right.bbox = UnionBboxes(rightBboxes) // 合并bounds
	// fmt.Println("right bboxes'count:", len(right.bboxes), " ids'count:", len(right.ids))
	return
}

// 合并bbox数组
func UnionBboxes(bboxes []base.Rect2D) (bbox base.Rect2D) {
	bbox.Init()
	for _, v := range bboxes {
		bbox.Union(v)
	}
	return
}

// 查询
// 从根节点开始，判断相交就往下找，直到叶子节点
func (this *RTreeNode) Query(bbox base.Rect2D) (ids []int) {
	// fmt.Println("RTreeNode.Query(),bbox:", bbox)

	if this.bbox.IsIntersect(bbox) {
		if this.isLeaf {
			for i, v := range this.bboxes {
				if bbox.IsIntersect(v) {
					ids = append(ids, this.ids[i])
				}
			}
		} else {
			for _, v := range this.nodes {
				if bbox.IsIntersect(v.bbox) {
					ids = append(ids, v.Query(bbox)...)
				}
			}
		}
	}
	// fmt.Println("query ids count:", len(ids))
	return
}

func (this *RTreeNode) Clear() {
	this.ids = this.ids[0:0]
	this.bboxes = this.bboxes[0:0]
	for _, v := range this.nodes {
		v.Clear()
	}
}
