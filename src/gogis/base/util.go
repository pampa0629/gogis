package base

import (
	"fmt"
	"math"
	"time"
)

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
	fmt.Println(dosome, " time: ", seconds, "millisecond")
}

// ================================================== //

// 计算得到点串的bounds
func ComputeBounds(points []Point2D) Rect2D {
	var bbox Rect2D
	bbox.Init()
	for _, pnt := range points {
		bbox.Min.X = math.Min(bbox.Min.X, pnt.X)
		bbox.Min.Y = math.Min(bbox.Min.Y, pnt.Y)
		bbox.Max.X = math.Max(bbox.Max.X, pnt.X)
		bbox.Max.Y = math.Max(bbox.Max.Y, pnt.Y)
	}
	return bbox
}

// ================================================== //

// todo 这个函数得换个包放置
// 根据bbox和对象数量，计算缓存的最小最大合适层级
// 再小的层级没有必要（图片上的显示范围太小）；再大的层级则瓦片上对象太稀疏
func CalcMinMaxLevels(bbox Rect2D, geoCount int) (minLevel, maxLevel int) {
	fmt.Println("bbox:", bbox)
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
// // todo
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
func RemoveRepByMap(ids []int) []int {
	// result := []int{}         //存放返回的不重复切片
	result := make([]int, 0, len(ids))
	tempMap := map[int]byte{} // 存放不重复主键
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
// 返回最大值对应的序号
func Max(values []float64) (no int) {
	maxValue := values[0]
	no = 0
	for i := 1; i < len(values); i++ {
		if values[i] > maxValue {
			maxValue = values[i]
			no = i
		}
	}
	return
}
