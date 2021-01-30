package base

import (
	"strings"
)

// 九交模型定义
// 内部：Interior；边界：Boundary；外部：Exterior
// 点：		内部：自身；边界：没有；外部：其它
// 线：		内部：自身（不含端点）；边界：端点；外部：其它
// 面（环）：内部：自身（不含边界）；边界：边线；外部：其它
// 维度(Dim)：-1：没有交集；0：交集为点；1：交集为线；2：交集为面
// T：有交集（0/1/2均可）；F：没有交集（-1）；*：T/F/0/1/2均可；
// 如：A relate B --> "1010F0212"
// A/B  I  B  E
// I    1  0  1
// B    0  F  0
// E    2  1  2

// 空间关系算子，JTS定义；注：每一个算子在面对不同点线面组合时，可能有不同的定义
// 基本参考：http://www.whudj.cn/?p=778 ，但Contains和Within的定义略有区别

// 相等(Equals)：	A和B拓扑上相等。 [T*F**FFF*] Dim也应相等
// 相离(Disjoint)：	A和B没有共有的交集。[FF*FF****] 即不相交
// 相交(Intersects)：A和B至少有一个共有点（区别于相离）,Not Disjoint
// 					有四种情况皆可：[T********]  [*T*******] [***T*****] [****T****]
// 接触(Touches)：	A和B有交集（内部与边界，边界与边界），但内部不能有交集。两者不能同时为点
// 					包括三种情况：[FT*******] [F**T*****] [F***T****]
// 穿越(Crosses)：	A和B内部有交集，且交集的维度比A和B最大维度要小。只适合线&线、线&面
// 					线&线: [0********] 线&面 [T*T******]
// 重叠(Overlaps)：	A和B的维度相同，有交集，也有部分不相交，交集的维度等于A和B的维度
// 					点&点，面&面：[T*T***T**]; 线&线：[1*T***T**]
// 包含(Contains)：	B都在A内部，与Within是倒置关系 [T***F*FF*]
// 					与Cover的区别在于：A Cover B, B的边界可能与A的边界有交集;
// 									  但A Contains B, B的边界不能与A的边界有交集
// 					注：这一点的定义可能和其他软件不同
// 在内(Within)：	A都在B的内部。[T*F*FF***]
// 覆盖(Covers)：	A覆盖B；A的外部与B没有交集 [******FF*] 与CoveredBy是倒置关系
// 					和Contain的区别在于：Cover允许两者的边界有交集，或点在边界上
// 被覆盖(CoveredBy)：A被B覆盖 [**F**F***]
// 在xx之上(On): 	等于CoveredBy, 一般用于点和线/面的关系判断

// 空间查询模式定义，详细解释请参考base中的algorithm.go
type SpatialMode string

// 可以a/b顺序互换的包括: Intersects,Disjoint,Equal,Overlap,Touch,Cross(线线)
// 互为逆运算的包括： Contains,Within; Cover,CoveredBy

const (
	Undefined      SpatialMode = ""               // 未定义, 默认等于Intersects
	Equals         SpatialMode = "Equals"         // 相等
	Disjoint       SpatialMode = "Disjoint"       // 相离
	Intersects     SpatialMode = "Intersects"     // 相交
	Touches        SpatialMode = "Touches"        // 接触
	Crosses        SpatialMode = "Crosses"        // 穿越
	Within         SpatialMode = "Within"         // 在内（与包含相反）
	Contains       SpatialMode = "Contains"       // 包含
	Overlaps       SpatialMode = "Overlaps"       // 重叠
	Covers         SpatialMode = "Covers"         // 覆盖
	CoveredBy      SpatialMode = "CoveredBy"      // 被覆盖
	BBoxIntersects SpatialMode = "BBoxIntersects" // 外接矩形相交 == QueryByBounds
	// todo 支持用户自定义，九交模型矩阵 DE-9IM: Dimensionally Extended 9-Intersection model
	// [*********] 0/1/2/T/F/*
)

// 判断该mode是否和bbox有关联
func (mode SpatialMode) IsDisjoint() (disjoint bool) {
	switch mode {
	case Disjoint:
		disjoint = true
	case Undefined, Equals, Intersects,
		Touches, Crosses, Within,
		Contains, Overlaps, Covers,
		CoveredBy, BBoxIntersects:
		disjoint = false
	default:
		var dm D9IM
		dm.Init(string(mode))
		disjoint = dm.IsDisjoint()
	}
	return
}

// ========================================================== //

type IBE int

const (
	I IBE = 0 // 内部
	B IBE = 1 // 边界
	E IBE = 2 // 外部
)

// DE-9IM: Dimensionally Extended 9-Intersection model
type D9IM struct {
	dm [][]byte // 0/1/2/F/T/*
}

func (this *D9IM) Init(str string) bool {
	str = strings.TrimSpace(str)
	str = strings.Trim(str, "[](){}")
	if len(str) == 9 {
		this.dm = make([][]byte, 3)
		for i, _ := range this.dm {
			this.dm[i] = make([]byte, 3)
			for j, _ := range this.dm[i] {
				if isValid(str[i*3+j]) {
					this.dm[i][j] = str[i*3+j]
				} else {
					return false
				}
			}
		}
		return true
	}
	return false
	// fmt.Println("dm:", this)
}

func isValid(v byte) bool {
	switch v {
	case '0', '1', '2', 'F', 'T', '*':
		return true
	default:
		return false
	}
}

// 得到一个具体
func (this *D9IM) Get(a, b IBE) byte {
	return this.dm[a][b]
}

func (this *D9IM) Set(a, b IBE, v byte) {
	this.dm[a][b] = v
}

// 必须每一个都符合im的要求，才返回true
func (this *D9IM) MatchIM(im D9IM) bool {
	for i, v := range this.dm {
		for ii, _ := range v {
			if !match(this.dm[i][ii], im.dm[i][ii]) {
				return false
			}
		}
	}
	return true
}

func (this *D9IM) Match(str string) bool {
	var im D9IM
	if im.Init(str) {
		return this.MatchIM(im)
	}
	return false
}

// 看是否匹配
// 0/1/2/F/T/*
func match(act, req byte) bool {
	switch req {
	case '*':
		return true
	case 'T':
		return act == 'T' || act == '0' || act == '1' || act == '2'
	case '2', '1', '0', 'F':
		return act == req
	}
	return false
}

// 转置矩阵 [0,1,2,3,4,5,6,7,8] 变为 [0,3,6,1,4,7,2,5,8]
func (this *D9IM) Invert() {
	for i, v := range this.dm {
		for j, _ := range v {
			this.dm[i][j], this.dm[j][i] = this.dm[j][i], this.dm[i][j]
		}
	}
}

func (this *D9IM) String() (out string) {
	for _, v := range this.dm {
		for _, vv := range v {
			out += string(vv)
		}
	}
	return
}

// 满足 [FF*FF****] 即可
func (this *D9IM) IsDisjoint() bool {
	if this.dm[0][0] == 'F' && this.dm[0][1] == 'F' &&
		this.dm[1][0] == 'F' && this.dm[1][1] == 'F' {
		return true
	}
	return false
}

// ========================================================== //

// 空间分析：
// 缓冲区分析（Buffer）	包含所有的点在一个指定距离内的多边形和多多边形。（更多信息请查阅《第五缓冲区计算机》第七页）
// 凸壳分析（ConvexHull）	包含几何形体的所有点的最小凸壳多边形（外包多边形）
// 交叉分析（Intersection）	交叉操作就是多边形AB中所有共同点的集合。
// 联合分析（Union）	AB的联合操作就是AB所有点的集合。
// 差异分析（Difference）	AB形状的差异分析就是A里有B里没有的所有点的集合。
// 对称差异分析（SymDifference）	AB形状的对称差异分析就是位于A中或者B中但不同时在AB中的所有点的集合
