package index

import (
	"encoding/binary"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"io"
	"math"
)

// 每个节点直接管辖的对象（子节点或id）个数的最大最小值
const RTREE_MIN_OBJ_COUNT = 10000
const RTREE_MAX_OBJ_COUNT = 50000

// 每个节点直接管辖的对象（子节点或id）个数
var RTREE_OBJ_COUNT int

// R树索引
type RTreeIndex struct {
	RTreeNode       // 根节点
	idCount   int64 // 索引所管辖的id数量
}

func (this *RTreeIndex) Type() SpatialIndexType {
	return TypeRTreeIndex
}

func (this *RTreeIndex) Init(bbox base.Rect2D, num int64) {
	this.RTreeNode.init(bbox)
	this.idCount = num
	maxCountPerNode := num / 16 // 为啥除以这个数，我也是不知道
	RTREE_OBJ_COUNT = base.IntMax(base.IntMin(int(maxCountPerNode), RTREE_MAX_OBJ_COUNT), RTREE_MIN_OBJ_COUNT)
	// this.objCount = RTREE_OBJ_COUNT
	fmt.Println("rtree node obj count:", RTREE_OBJ_COUNT)
}

func (this *RTreeIndex) AddGeos(geometrys []geometry.Geometry) {
	for _, geo := range geometrys {
		// fmt.Println("start deal one geo, id:", i)
		this.AddOne(geo.GetBounds(), geo.GetID())
		// fmt.Println("end  deal one geo, id:", i)
		// this.WholeString()
	}
}

func (this *RTreeIndex) AddGeo(geo geometry.Geometry) {
	this.AddOne(geo.GetBounds(), geo.GetID())
}

func (this *RTreeIndex) Load(r io.Reader) {
	tr := base.NewTimeRecorder()

	// this.LoadHeader(r)
	binary.Read(r, binary.LittleEndian, &this.idCount) // 8
	this.RTreeNode.load(r)

	tr.Output("load rtree index")
}

func (this *RTreeIndex) LoadHeader(r io.Reader) {
	var headerLength, typeIndex, version int32         // 记录头100字节
	binary.Read(r, binary.LittleEndian, &headerLength) // 4
	binary.Read(r, binary.LittleEndian, &typeIndex)    // 4
	binary.Read(r, binary.LittleEndian, &version)      // 4
	// binary.Read(r, binary.LittleEndian, &this.objCount) // 4
	binary.Read(r, binary.LittleEndian, &this.idCount) // 8

	space := make([]byte, headerLength-20)
	binary.Read(r, binary.LittleEndian, space)
}

// 保存
func (this *RTreeIndex) Save(w io.Writer) {
	binary.Write(w, binary.LittleEndian, this.idCount) // 8
	// this.saveHeader(w)
	this.RTreeNode.save(w)
}

// 保存索引头
// func (this *RTreeIndex) saveHeader(w io.Writer) {
// 	headerLength := int32(100)                           // 记录头100字节
// 	binary.Write(w, binary.LittleEndian, headerLength)   // 4
// 	binary.Write(w, binary.LittleEndian, TypeRTreeIndex) // 4
// 	version := int32(1)
// 	binary.Write(w, binary.LittleEndian, version) // 4
// 	// binary.Write(w, binary.LittleEndian, this.objCount) // 4
// 	binary.Write(w, binary.LittleEndian, this.idCount) // 8

// 	binary.Write(w, binary.LittleEndian, [80]byte{}) //
// }

// 自我检查，发现构建好的索引是否符合基本规则
// 可能的问题包括：
// level 没有按照0-1-2递增
// 子节点的父节点不是自己
// 叶子节点有nodes，非叶子节点有ids
//     所管理数量超过 RTREE_OBJ_COUNT
// 所管理范围不等于bbox
// id不重复，且总数等于输入值
func (this *RTreeIndex) Check() bool {
	if !this.RTreeNode.checkLevel() {
		return false
	}
	if !this.RTreeNode.checkParent() {
		return false
	}
	if !this.RTreeNode.checkManager() {
		return false
	}
	if !this.RTreeNode.checkBbox() {
		return false
	}

	ids := map[int64]byte{} // 存放不重复主键
	if !this.RTreeNode.checkID(&ids) {
		return false
	}
	// 总数也必须相等
	if int64(len(ids)) != this.idCount {
		return false
	}

	return true
}

// R树的一个节点
type RTreeNode struct {
	level  int32       // 层级
	parent *RTreeNode  // 父节点，根节点的父节点为nil
	bbox   base.Rect2D // 本节点的bounds
	isLeaf bool        // 是否为叶子节点；如果是叶子节点，则没有子节点，存储对象（id和bbox）数组；非子节点时，存储子节点数组

	nodes []*RTreeNode // 子节点

	ids    []int64       // 对象id
	bboxes []base.Rect2D // 对象bounds
}

// 保存节点
func (this *RTreeNode) save(w io.Writer) {
	binary.Write(w, binary.LittleEndian, this.level)
	binary.Write(w, binary.LittleEndian, base.Bool2Int32(this.isLeaf))
	binary.Write(w, binary.LittleEndian, this.bbox)
	if this.isLeaf {
		binary.Write(w, binary.LittleEndian, int32(len(this.ids)))
		binary.Write(w, binary.LittleEndian, this.ids)
		binary.Write(w, binary.LittleEndian, this.bboxes)
	} else {
		binary.Write(w, binary.LittleEndian, int32(len(this.nodes)))
		for _, node := range this.nodes {
			node.save(w)
		}
	}
}

func (this *RTreeNode) load(r io.Reader) {
	binary.Read(r, binary.LittleEndian, &this.level)
	var isLeaf int32
	binary.Read(r, binary.LittleEndian, &isLeaf)
	this.isLeaf = isLeaf != 0
	binary.Read(r, binary.LittleEndian, &this.bbox)
	var count int32
	binary.Read(r, binary.LittleEndian, &count)
	if this.isLeaf {
		this.ids = make([]int64, count)
		binary.Read(r, binary.LittleEndian, this.ids)
		this.bboxes = make([]base.Rect2D, count)
		binary.Read(r, binary.LittleEndian, this.bboxes)
	} else {
		this.nodes = make([]*RTreeNode, count)
		for i, _ := range this.nodes {
			node := new(RTreeNode)
			node.load(r)
			node.parent = this
			this.nodes[i] = node
		}
	}
}

// 检查level层级是否正确
func (this *RTreeNode) checkLevel() bool {
	// 父节点比我小1
	if this.parent != nil {
		if this.parent.level != this.level-1 {
			return false
		}
	}
	// 子节点比我大1
	if !this.isLeaf {
		for _, v := range this.nodes {
			if v.level != this.level+1 {
				return false
			}
			// 迭代检查
			if !v.checkLevel() {
				return false
			}
		}
	}
	return true
}

// 检查父子节点关系
func (this *RTreeNode) checkParent() bool {
	// 子节点的父节点必须是自己
	if !this.isLeaf {
		for _, v := range this.nodes {
			if v.parent != this {
				return false
			}
			if !v.checkParent() {
				return false
			}
		}
	}

	return true
}

// 检查管理范围是否正确
// 叶子节点有nodes，非叶子节点有ids
// 所管理数量超过 RTREE_OBJ_COUNT
func (this *RTreeNode) checkManager() bool {
	if this.isLeaf {
		if len(this.nodes) > 0 {
			return false
		}
		// if len(this.ids) > objCount {
		// 	return false
		// }
	}

	if !this.isLeaf {
		bboxCount := len(this.bboxes)
		idCount := len(this.ids)
		if bboxCount > 0 || idCount > 0 || bboxCount != idCount {
			return false
		}
		// if len(this.nodes) > objCount {
		// 	return false
		// }
		for _, v := range this.nodes {
			if !v.checkManager() {
				return false
			}
		}
	}

	return true
}

// 所管理范围不等于bbox
func (this *RTreeNode) checkBbox() bool {
	if this.isLeaf {
		if base.UnionBounds(this.bboxes) != this.bbox {
			return false
		}
	}

	if !this.isLeaf {
		var bbox base.Rect2D
		bbox.Init()
		for _, v := range this.nodes {
			bbox.Union(v.bbox)
		}
		if bbox != this.bbox {
			return false
		}

		for _, v := range this.nodes {
			if !v.checkBbox() {
				return false
			}
		}
	}
	return true
}

// id不重复，且总数等于输入值
func (this *RTreeNode) checkID(ids *map[int64]byte) bool {
	if this.isLeaf {
		for _, id := range this.ids {
			idCount := len(*ids)
			(*ids)[id] = 0            // 如果 id 已经存在，这里就添加不上去，从而导致ids的长度不增加
			if len(*ids) == idCount { // id重复了，返回false
				return false
			}
		}
	} else {
		for _, node := range this.nodes {
			if !node.checkID(ids) {
				return false
			}
		}
	}

	return true
}

func (this *RTreeNode) space() (msg string) {
	for i := int32(0); i < this.level; i++ {
		msg += "    "
	}
	return
}

func (this *RTreeNode) string() {
	fmt.Println(this.space(), "level:", this.level, ", isLeaf:", this.isLeaf, ", bbox:", this.bbox, ", nodes'count:", len(this.nodes), ", ids's count:", len(this.ids))
	// fmt.Println("")
}
func (this *RTreeNode) wholeString() {
	this.string()
	if this.isLeaf {
		for i, _ := range this.ids {
			fmt.Println(this.space(), "    id:", this.ids[i], "bbox:", this.bboxes[i])
		}

	} else {
		for _, v := range this.nodes {
			v.wholeString()
		}
	}
}

func (this *RTreeNode) init(bbox base.Rect2D) {
	this.level = 0
	this.parent = nil
	this.bbox = bbox
	this.isLeaf = true
}

func (this *RTreeNode) AddOne(bbox base.Rect2D, id int64) {
	// fmt.Println("RTreeNode.addOne()", bbox, id)
	// this.String()

	// 过程：先看自己是否叶子节点；
	//     若是，则给自己添加对象，添加后，若超过最大个数，则引发自己分裂为两个叶子节点；
	//         怎么个分裂法，得看分裂后，两个节点的范围尽可能的分开，且面积总和最小
	//         之后还得看父节点所管理的子节点个数是否超过最大个数，若是，则引发父节点分裂；往上以此类推
	//     若不是叶子节点，则往下找一个最合适的节点，以此类推，直到找到最合适的叶子节点来存放（之后有可能引发自己的分裂）
	//         最合适的标准为：加入后，使得节点范围的面积增加值最小

	this.bbox.Union(bbox) // 自己的范围要扩大

	if this.isLeaf {
		this.ids = append(this.ids, id)
		this.bboxes = append(this.bboxes, bbox)
		this.bbox.Union(bbox)
		if len(this.ids) >= RTREE_OBJ_COUNT {
			this.split()
		}
	} else {
		node := this.findBestChild(bbox)
		node.AddOne(bbox, id)
	}
}

// 找到最合适的子节点，以便往该子节点中增加对象
// 标准为：对象加入后，使得节点范围的面积增加值最小
func (this *RTreeNode) findBestChild(bbox base.Rect2D) *RTreeNode {
	// fmt.Println("RTreeNode.findBestChild()", bbox)
	// this.String()

	minMoreArea := math.MaxFloat64 // 新增加的面积，一开始设置为最大值
	minNode := (*RTreeNode)(nil)
	for _, node := range this.nodes {
		newBbox := node.bbox
		newBbox.Union(bbox)
		moreArea := newBbox.Area() - node.bbox.Area()
		if moreArea < minMoreArea {
			minMoreArea = moreArea
			minNode = node
		}
	}
	return minNode
}

// 节点分裂为两个同级别的节点
// 标准为：两个节点的范围应尽可能分离，且面积总和最小
// 分裂后，若导致父节点所管理的子节点数量超过最大值，父节点也得进一步分裂；迭代往复
// 注意：本节点可能为叶子节点，也可能是中间节点，还有可能是根节点，根节点也有可能同时就是叶子节点；得分别处理
func (this *RTreeNode) split() {
	// fmt.Println("RTreeNode.split()")
	// this.String()

	// 若自己就是根节点，则自己不能动，把从自己分裂出来的两个节点纳入自己管辖
	leftNode := (*RTreeNode)(nil)
	rightNode := (*RTreeNode)(nil)

	if this.isLeaf {
		// 叶子节点分裂，把自己变为两个，划分好对象
		leftNode, rightNode = this.splitLeaf()
	} else {
		// 中间节点分裂，是把自己的子节点分成两拨，分别归到新生成的两个同级节点中
		leftNode, rightNode = this.splitNoLeaf()
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
		// 根节点分裂，会导致下面所有节点的level都要加1
		leftNode.addChildNodeLevel()
		rightNode.addChildNodeLevel()
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
		if len(this.parent.nodes) >= RTREE_OBJ_COUNT {
			// fmt.Println("before parent split")
			// this.String()
			// this.getRoot().WholeString()
			this.parent.split()
		}
	}

	// this.getRoot().WholeString()
	return
}

// 得到根节点
func (this *RTreeNode) getRoot() (root *RTreeNode) {
	root = this
	for {
		if root.parent == nil {
			break
		}
		root = root.parent
	}
	return
}

// 给下级节点增加level值，并迭代下去
func (this *RTreeNode) addChildNodeLevel() {
	for _, v := range this.nodes {
		v.level += 1
		v.addChildNodeLevel()
	}
}

// 分裂非叶子节点
func (this *RTreeNode) splitNoLeaf() (left, right *RTreeNode) {
	// fmt.Println("RTreeNode.SplitNoLeaf()")
	// this.String()
	leftNodes, rightNodes, leftBbox, rightBbox := splitNodes(this.nodes)
	left = createNoLeafChildNode(leftNodes, leftBbox)
	right = createNoLeafChildNode(rightNodes, rightBbox)
	return
}

// 创建非叶子的子节点
func createNoLeafChildNode(nodes []*RTreeNode, bbox base.Rect2D) *RTreeNode {
	childNode := new(RTreeNode)
	childNode.isLeaf = false
	childNode.nodes = nodes
	childNode.bbox = bbox
	for _, v := range childNode.nodes {
		v.parent = childNode
	}
	return childNode
}

// 把节点数组分为两个
func splitNodes(nodes []*RTreeNode) (leftNodes, rightNodes []*RTreeNode, leftBbox, rightBbox base.Rect2D) {
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
func splitBoxes(bboxes []base.Rect2D, ids []int64) (leftBboxes []base.Rect2D, leftIds []int64, rightBboxes []base.Rect2D, rightIds []int64) {
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
func findSeedBoxes(bboxes []base.Rect2D, ids []int64) (leftBox base.Rect2D, leftId int64, rightBox base.Rect2D, rightId int64) {
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
	wholeBbox := base.UnionBounds(bboxes)
	leftPos := findNeareastPoint(wholeBbox.Min, centerPnts)
	rightPos := findNeareastPoint(wholeBbox.Max, centerPnts)
	if leftPos == rightPos {
		fmt.Println("出现特例了,left pos == right pos, value is:", leftPos)
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
func findSeedBoxes2(bboxes []base.Rect2D, ids []int64) (leftBox base.Rect2D, leftId int64, rightBox base.Rect2D, rightId int64) {
	// fmt.Println("RTreeNode.findSeedBoxes2(), boxes's count:", len(bboxes), ",ids's count:", len(ids))

	count := len(ids)
	maxDist := make([]float64, count)
	maxId := make([]int64, count)
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
	ids := make([]int64, len(nodes))
	for i, v := range nodes {
		bboxes[i] = v.bbox
		ids[i] = int64(i)
	}
	_, leftId, _, rightId := findSeedBoxes(bboxes, ids)
	leftNode = nodes[leftId]
	rightNode = nodes[rightId]
	return
}

// 分裂叶子节点
func (this *RTreeNode) splitLeaf() (left, right *RTreeNode) {
	// fmt.Println("RTreeNode.SplitLeaf()")
	// this.String()

	leftBboxes, leftIds, rightBboxes, rightIds := splitBoxes(this.bboxes, this.ids)

	left = new(RTreeNode)
	left.isLeaf = true
	left.bboxes = leftBboxes
	left.ids = leftIds
	left.bbox = base.UnionBounds(leftBboxes) // 合并bounds
	// fmt.Println("left bboxes'count:", len(left.bboxes), " ids'count:", len(left.ids))

	right = new(RTreeNode)
	right.isLeaf = true
	right.bboxes = rightBboxes
	right.ids = rightIds
	right.bbox = base.UnionBounds(rightBboxes) // 合并bounds
	// fmt.Println("right bboxes'count:", len(right.bboxes), " ids'count:", len(right.ids))
	return
}

// 查询
// 从根节点开始，判断相交就往下找，直到叶子节点
func (this *RTreeNode) Query(bbox base.Rect2D) (ids []int64) {
	// fmt.Println("RTreeNode.Query(),bbox:", bbox)

	if this.bbox.IsIntersect(bbox) {
		if this.isLeaf {
			// ids = append(ids, this.ids...) 模糊查找
			// 精确查找
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
	this.nodes = this.nodes[0:0]
}
