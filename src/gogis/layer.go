package gogis

import (
	"sync"
)

type Layer struct {
	Shp *ShapeFile
}

func NewLayer(shp *ShapeFile) *Layer {
	layer := new(Layer)
	layer.Shp = shp
	return layer
}

// 一次性绘制的对象个数
const ONE_DRAW_COUNT = 100000

func (this *Layer) Draw(canvas *Canvas) int {
	ids := this.Shp.Query(canvas.params.GetBounds())
	// fmt.Println("ids count:", len(ids))

	forcount := (int)(len(ids)/ONE_DRAW_COUNT) + 1
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < forcount; i++ {
		wg.Add(1)
		go this.drawBatch(i*ONE_DRAW_COUNT, ids, canvas, wg)
	}
	wg.Wait()
	return len(ids)
}

func (this *Layer) drawBatch(num int, ids []int, canvas *Canvas, wg *sync.WaitGroup) {
	// fmt.Println("begin drawBatch:", num)
	for i := 0; i < ONE_DRAW_COUNT && num < len(ids); i++ {
		// line := ChangePolyline(this.Shp.geometrys[ids[num]], canvas.params)
		if this.Shp.geoPyms[10][ids[num]] != nil {
			line := ChangePolyline(this.Shp.geoPyms[10][ids[num]], canvas.params)
			canvas.DrawPolyline(line)
		}
		num++
	}
	// png.Encode(f, img)
	defer wg.Done()
}

// 把 shape格式（浮点数）的对象，转化为绘制格式（整数）的对象，方便后续绘制
func ChangePolyline(polyline *shpPolyline, params CoordParams) *IntPolyline {
	// fmt.Println("ChangePolyline: ", polyline)
	var intPolyline = new(IntPolyline)
	intPolyline.numParts = (int)(polyline.numParts)
	intPolyline.points = make([][]Point, intPolyline.numParts)
	pos := 0
	for i := 0; i < (int)(polyline.numParts); i++ {
		pntCount := 0
		if i < (int)(polyline.numParts)-1 {
			pntCount = (int)(polyline.parts[i+1] - polyline.parts[i])
		} else {
			pntCount = (int)(polyline.numPoints - polyline.parts[i])
		}
		intPolyline.points[i] = make([]Point, pntCount)

		for j := 0; j < pntCount; j++ {
			intPolyline.points[i][j] = params.Forward((Point2D)(polyline.points[pos]))
			// intPolyline.points[i][j].X = (int)(scaleX * (polyline.points[pos].X - xmin))
			// intPolyline.points[i][j].Y = dy - (int)(scaleY*(polyline.points[pos].Y-ymin))
			pos++
		}
	}
	// fmt.Println("before geometry:", polyline)
	// fmt.Println("after geometry:", intPolyline)
	return intPolyline
}
