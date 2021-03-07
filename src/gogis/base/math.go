package base

import (
	"math"
	"strings"
	"time"
)

// 求绝对值
func Abs(x int) int {
	if x >= 0 {
		return x
	}
	return -x
}

// 四舍五入
func Round(x float64) int {
	return int(math.Floor(x + 0.5))
}

// 求a的n次方
func Power(a, n int) int {
	return int(math.Pow(float64(a), float64(n)))
}

// 求最小值
func IntMin(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// 求最大值
func IntMax(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func Int32Max(x, y int32) int32 {
	if x > y {
		return x
	}
	return y
}

// 返回最大值对应的序号
func Max(values []float64) (no int) {
	maxValue := values[0]
	for i := 1; i < len(values); i++ {
		if values[i] > maxValue {
			maxValue = values[i]
			no = i
		}
	}
	return
}

// 得到最大值和最小值
func GetExtreme(values []int) (max, min int) {
	max, min = values[0], values[0]
	for i := 1; i < len(values); i++ {
		if values[i] > max {
			max = values[i]
		} else if values[i] < min {
			min = values[i]
		}
	}
	return
}

func IsMatchBool(value1 bool, op string, value2 bool) bool {
	switch op {
	case "=":
		return value1 == value2
	case "!=":
		return value1 != value2
	}
	return false
}

func IsMatchInt(value1 int, op string, value2 int) bool {
	switch op {
	case "=":
		return value1 == value2
	case "!=":
		return value1 != value2
	case ">":
		return value1 > value2
	case "<":
		return value1 < value2
	case ">=":
		return value1 >= value2
	case "<=":
		return value1 <= value2
	}
	return false
}

const FLOAT_ZERO = 10e-10 // 定义接近0的极小值
// const FLOAT_ZERO = math.SmallestNonzeroFloat64 * 10

// 浮点数相等比较
func IsEqual(value1 float64, value2 float64) bool {
	if math.Abs(value1-value2) < FLOAT_ZERO {
		return true
	}
	return false
}

// 浮点数大于等于
func IsBigEqual(value1 float64, value2 float64) bool {
	if value1 > value2 || math.Abs(value1-value2) < FLOAT_ZERO {
		return true
	}
	return false
}

// 浮点数小于等于
func IsSmallEqual(value1 float64, value2 float64) bool {
	if value1 < value2 || math.Abs(value1-value2) < FLOAT_ZERO {
		return true
	}
	return false
}

func IsMatchFloat(value1 float64, op string, value2 float64) bool {
	switch op {
	case "=":
		return IsEqual(value1, value2)
	case "!=":
		return !IsEqual(value1, value2)
	case ">":
		return value1 > value2
	case "<":
		return value1 < value2
	case ">=":
		return value1 >= value2
	case "<=":
		return value1 <= value2
	}
	return false
}

func IsMatchString(value1 string, op string, value2 string) bool {
	switch op {
	case "=":
		return strings.EqualFold(value1, value2)
	case "!=":
		return !strings.EqualFold(value1, value2)
		// 感觉没啥用，先封起来
		// case ">":
		// 	return value1 > value2
		// case "<":
		// 	return value1 < value2
		// case ">=":
		// 	return value1 >= value2
		// case "<=":
		// 	return value1 <= value2
	}
	return false
}

func IsMatchTime(value1 time.Time, op string, value2 time.Time) bool {
	switch op {
	case "=":
		return value1.Equal(value2)
	case "!=":
		return !value1.Equal(value2)
	case ">":
		return value1.After(value2)
	case "<":
		return value1.Before(value2)
	case ">=":
		return value1.Equal(value2) || value1.After(value2)
	case "<=":
		return value1.Before(value2) || value1.After(value2)
	}
	return false
}

// 计算两点距离的平方
func DistanceSquare(x0, y0, x1, y1 float64) float64 {
	return math.Pow((x0-x1), 2) + math.Pow((y0-y1), 2)
}

// ===================================================== //

// 支持： sort.Sort(base.Int64s([]int64))
type Int64s []int64

//Len()
func (s Int64s) Len() int {
	return len(s)
}

//Less():成绩将有低到高排序
func (s Int64s) Less(i, j int) bool {
	return s[i] < s[j]
}

//Swap()
func (s Int64s) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// 支持： sort.Sort(base.Int32s([]int32))
type Int32s []int32

//Len()
func (s Int32s) Len() int {
	return len(s)
}

//Less():成绩将有低到高排序
func (s Int32s) Less(i, j int) bool {
	return s[i] < s[j]
}

//Swap()
func (s Int32s) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// ===================================================== //

// p1,p2,p3三个点，判断p3在p1p2向量的左边还是右边，左右跟向量的方向有关，
// 如果是p1p2的方向，那么就是对|p1,p2,p3|进行叉积计算，
// 根据右手法则，如果计算的答案大于0，就是左侧，小于0就是右侧，等于0就是在直线上。
// 叉乘；若结果为0，说明共线
func CrossX(p1, p2, p Point2D) float64 {
	//两点p1(x1,y1),p2(x2,y2),判断点p(x,y)
	// (y1 – y2) * x + (x2 – x1) * y + x1 * y2 – x2 * y1
	return (p1.Y-p2.Y)*p.X + (p2.X-p1.X)*p.Y + p1.X*p2.Y - p2.X*p1.Y
}
