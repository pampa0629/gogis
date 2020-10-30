package mapping

import (
	"gogis/data"
	"gogis/geometry"
	"sync"
)

type Layer struct {
	Name   string          // 图层名
	feaset data.Featureset // 干活用的
	Params data.ConnParams `json:"ConnParams"` // 存储和打开地图文档时用的
}

func NewLayer(feaset data.Featureset) *Layer {
	layer := new(Layer)
	layer.feaset = feaset
	store := feaset.GetStore()
	if store != nil {
		layer.Params = store.GetConnParams()
		layer.Params["name"] = feaset.GetName()
		layer.Name = layer.Params["name"] // 默认图层名 等于 数据集名
	}

	return layer
}

// 一次性绘制的对象个数
const ONE_DRAW_COUNT = 100000

func (this *Layer) Draw(canvas *Canvas) int {
	feait := this.feaset.QueryByBounds(canvas.params.GetBounds())
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for {
		features, ok := feait.BatchNext(ONE_DRAW_COUNT)
		if ok {
			wg.Add(1)
			go this.drawBatch(features, canvas, wg)
		} else {
			break
		}
	}
	wg.Wait()

	return feait.Count()
}

func (this *Layer) drawBatch(features []data.Feature, canvas *Canvas, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, v := range features {
		switch v.Geo.Type() {
		case geometry.TGeoPolyline:
			line := ChangePolyline(v.Geo.(*geometry.GeoPolyline), canvas.params)
			//  todo 用矢量金字塔进行绘制
			// if this.Shp.geoPyms[4][ids[num]] != nil {
			// line := ChangePolyline(this.Shp.geoPyms[10][ids[num]], canvas.params)
			canvas.DrawPolyline(line)
		case geometry.TGeoPolygon:
			line := ChangePolygon(v.Geo.(*geometry.GeoPolygon), canvas.params)
			canvas.DrawPolyline(line)
		}
	}
}

// 把 空间对象，转化为绘制格式（整数）的对象，方便后续绘制
func ChangePolyline(polyline *geometry.GeoPolyline, params CoordParams) *IntPolyline {
	// fmt.Println("ChangePolyline: ", polyline)
	var intPolyline = new(IntPolyline)
	intPolyline.points = make([][]Point, len(polyline.Points))
	for i, v := range polyline.Points {
		intPolyline.points[i] = params.Forwards(v)
	}
	return intPolyline
}

// todo 未来再考虑转为面，以及 面的绘制
// 把 空间对象，转化为绘制格式（整数）的对象，方便后续绘制
func ChangePolygon(polygon *geometry.GeoPolygon, params CoordParams) *IntPolyline {
	// fmt.Println("ChangePolyline: ", polyline)
	var intPolyline = new(IntPolyline)
	intPolyline.points = make([][]Point, 0)
	for i, _ := range polygon.Points {
		for _, v := range polygon.Points[i] {
			pnts := params.Forwards(v)
			intPolyline.points = append(intPolyline.points, pnts)
		}
	}
	return intPolyline
}
