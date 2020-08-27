package base

import "math"

// 求绝对值
func Abs(x int) int {
	if x >= 0 {
		return x
	}
	return -x
}

func IntMin(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func IntMax(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// 计算两点距离的平方
func DistanceSquare(x0, y0, x1, y1 float64) float64 {
	return math.Pow((x0-x1), 2) + math.Pow((y0-y1), 2)
}
