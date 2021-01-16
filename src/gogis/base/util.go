// 不便归类的各种工具类功能

package base

import (
	"fmt"
	"math"
	"time"
)

// ================================================== //
// 时间记录，用来方便输出某个代码段的
type TimeRecorder struct {
	start int64
}

func NewTimeRecorder() (rt *TimeRecorder) {
	rt = new(TimeRecorder)
	rt.start = time.Now().UnixNano()
	return
}

func (this *TimeRecorder) Output(dosome string) {
	endTime := time.Now().UnixNano()
	seconds := float64((endTime - this.start) / 1e6)
	fmt.Println(dosome, "time: ", seconds, "millisecond", ", tick:", endTime-this.start)
	this.start = endTime // 支持连续输出
}

// ================================================== //
// 各类文件名的后缀，放在这里统一管理
const EXT_MAP_FILE = "gmp"           // 地图文档 Gogis MaP file
const EXT_SPATIAL_INDEX_FILE = "gix" // 空间索引文件，支持多种类型的空间索引存储 Gogis spatial IndeX file
const EXT_SQLITE_FILE = "gsf"        // 基于sqlite的空间文件引擎 Gogis Spatial File
const EXT_MAP_SERVICE_FILE = "gms"   // 地图服务文件 Gogis Map Service file

// ================================================== //

// todo 这个函数得换个包放置
// 根据bbox和对象数量，计算缓存的最小最大合适层级
// 再小的层级没有必要（图片上的显示范围太小）；再大的层级则瓦片上对象太稀疏
func CalcMinMaxLevels(bbox Rect2D, geoCount int64) (minLevel, maxLevel int32) {
	// fmt.Println("bbox:", bbox)
	minLevel = 0
	dis := 180.0
	dx := bbox.Dx()
	dy := bbox.Dy()
	// 地图长宽加一起，得够一个瓦片的长度，才值得从这个层级出图；不然就继续放大层级
	for dx+dy < dis {
		minLevel++
		dis /= 2.0
	}

	// 最大层级计算，要求每个瓦片的平均对象个数不少于 特定个数
	const GEO_MIN_COUNT = 10 // 考虑空白区域
	maxLevel = minLevel      // 从最小层级开始，后面每次瓦片数量*4
	for geoCount > GEO_MIN_COUNT {
		maxLevel++
		geoCount /= 4
	}

	return
}

// 计算特定层级，每个像素的经纬度长度；按照每个瓦片256*256计算
func CalcLevelDis(level int) float64 {
	dis := 360.0
	for level > 0 {
		level--
		dis /= 2.0
	}
	return dis / 256.0
}

// 计算并设置 web出图合适的 绘制参数params
// todo
// func SetParams(gmap *Map, nmap *Map, size int, row int, col int) {
// 	// 根据 row  和 col 修改 map的bbox
// 	// dx := gmap.BBox.Dx() / 4
// 	// dy := gmap.BBox.Dy() / 4
// 	// todo 1024 的修改
// 	change := 1024 / float64(size)
// 	scale := nmap.canvas.params.scale
// 	dx := float64(gmap.canvas.params.dx) / scale / change
// 	dy := float64(gmap.canvas.params.dy) / scale / change

// 	nmap.BBox = gmap.BBox
// 	nmap.BBox.Min.X += float64(col) * dx
// 	nmap.BBox.Max.X = nmap.BBox.Min.X + dx

// 	nmap.BBox.Max.Y -= float64(row) * dy
// 	nmap.BBox.Min.Y = nmap.BBox.Max.Y - dy
// }

// ================================================== //

// 去掉重复元素
func RemoveRepByMap(ids []int64) []int64 {
	// result := []int{}         //存放返回的不重复切片
	result := make([]int64, 0, len(ids))
	tempMap := map[int64]byte{} // 存放不重复主键
	for _, id := range ids {
		l := len(tempMap)
		tempMap[id] = 0 //当e存在于tempMap中时，再次添加是添加不进去的，，因为key不允许重复
		//如果上一行添加成功，那么长度发生变化且此时元素一定不重复
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, id) //当元素不重复时，将元素添加到切片result中
		}
	}
	return result
}

// ================================================== //

// 把一个数组尽可能均分为 n 份，返回数组的 数组
func SplitSlice32(s []int32, n int) (r [][]int32) {
	r = make([][]int32, n)
	count := len(s) / n // 每份分几个
	// 前面 n-1 份都相等
	for i := 0; i < n-1; i++ {
		r[i] = make([]int32, 0, count)
		r[i] = append(r[i], s[i*count:(i+1)*count]...)
	}
	// 剩下的都给最后一份
	r[n-1] = make([]int32, 0, len(s)-(n-1)*count)
	r[n-1] = append(r[n-1], s[(n-1)*count:]...)
	return
}

func SplitSlice64(s []int64, n int) (r [][]int64) {
	r = make([][]int64, n)
	count := len(s) / n // 每份分几个
	// 前面 n-1 份都相等
	for i := 0; i < n-1; i++ {
		r[i] = make([]int64, 0, count)
		r[i] = append(r[i], s[i*count:(i+1)*count]...)
	}
	// 剩下的都给最后一份
	r[n-1] = make([]int64, 0, len(s)-(n-1)*count)
	r[n-1] = append(r[n-1], s[(n-1)*count:]...)
	return
}

// ================================================ //
// 深度拷贝一个 map 或 切片
func DeepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}
		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy(v)
		}
		return newSlice
	}
	return value
}

// 计算夹角的角度；pnt0 为中间点
func Angle(pnt1, pnt0, pnt2 Point2D) float64 {
	x1 := pnt1.X - pnt0.X
	y1 := pnt1.Y - pnt0.Y
	x2 := pnt2.X - pnt0.X
	y2 := pnt2.Y - pnt0.Y

	x := x1*x2 + y1*y2
	y := x1*y2 - x2*y1
	radian := math.Acos(x / math.Sqrt(x*x+y*y))
	return radian * 180.0 / math.Pi
}

// ================================================================ //
func PrintError(msg string, err error) {
	if err != nil {
		fmt.Println(msg+" error:", err)
	}
}
