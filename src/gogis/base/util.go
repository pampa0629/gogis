package base

import (
	"fmt"
)

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
