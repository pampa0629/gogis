package draw

import (
	"gogis/base"
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

// go原生画布，采用gg和image绘制
// 左下角为坐标原点，X+ 向右，Y+向上
type Canvasgg struct {
	dc *gg.Context
	CoordParams
	style Style
}

func (this *Canvasgg) Clone() Canvas {
	var canvas = new(Canvasgg)
	canvas.CoordParams = this.CoordParams
	canvas.dc = gg.NewContext(canvas.dc.Width(), canvas.dc.Height())
	canvas.SetStyle(this.style)
	return canvas
}

// 初始化: 计算坐标转化参数，构造dc
func (this *Canvasgg) Init(bbox base.Rect2D, width, height int, data interface{}) {
	this.CoordParams.Init(bbox, float32(width), float32(height), false, Point{0, 0})
	if data == nil {
		this.dc = gg.NewContext(width, height)
	} else if img, ok := data.(*image.RGBA); ok {
		this.dc = gg.NewContextForRGBA(img)
	}
}

// func (this *Canvasgg) InitFromImage(bbox base.Rect2D, img *image.RGBA) {
// 	this.Params.Init(bbox, img.Rect.Size().X, img.Rect.Size().Y)
// 	// this.dc = gg.NewContext(width, height)
// 	this.dc = gg.NewContextForRGBA(img)
// 	// fmt.Println("dc:", this.dc)
// }

// 清空DC，为下次绘制做好准备
func (this *Canvasgg) Clear() {
	this.dc.Clear()
}

// todo
func (this *Canvasgg) Destroy() {
}

func (this *Canvasgg) GetImage() image.Image {
	return this.dc.Image()
}

func (this *Canvasgg) GetSize() (int, int) {
	return this.dc.Width(), this.dc.Height()
}

func (this *Canvasgg) SetStyle(style Style) {
	this.style = style
	// this.dc.SetFillStyle(gg.NewSolidPattern(style.FillColor))
	this.dc.SetFillColor(color.RGBA(style.FillColor))
	this.dc.SetStrokeColor(color.RGBA(style.LineColor))

	// this.dc.SetColor(style.LineColor)
	this.dc.SetLineWidth(style.LineWidth)
	this.dc.SetDash(style.LineDash...)
}

func (this *Canvasgg) SetTextColor(textColor color.RGBA) {
	pat := gg.NewSolidPattern(textColor)
	this.dc.SetFillStyle(pat)
}

// func (this *Canvasgg) Stroke() {
// 	this.dc.Stroke()
// }

func (this *Canvasgg) DrawImage(img image.Image, x, y float32) {
	this.dc.DrawImage(img, int(x), int(y))
}

func (this *Canvasgg) DrawPoint(pnt Point) {
	this.dc.DrawPoint(float64(pnt.X), float64(pnt.Y), 1)
	this.dc.Stroke()
}

func (this *Canvasgg) DrawLine(pnts []Point) {
	count := len(pnts)
	if count >= 2 {
		this.dc.MoveTo(float64(pnts[0].X), float64(pnts[0].Y))
		for i := 1; i < len(pnts)-1; i++ {
			this.dc.LineTo(float64(pnts[i].X), float64(pnts[i].Y))
		}
	}
	this.dc.Stroke()
}

// 绘制复杂线
// func (this *Canvasgg) DrawPolyPolyline(polyline *Polyline) {
// 	for _, v := range polyline.Points {
// 		this.DrawPolyline(v)
// 	}
// }

// 绘制复杂面（带洞）
// len必须大于1；[0] 是面，后面的都是洞；点的绕圈方向不论
func (this *Canvasgg) DrawPolyPolygon(polygon *Polygon) {
	polyCount := len(polygon.Points)
	if polyCount == 1 {
		// 简单多边形
		this.DrawPolygon(polygon.Points[0])
	} else if polyCount > 1 {
		// 先绘制后面的洞，再clip、mask一下，最后绘制面
		for i := 1; i < polyCount; i++ {
			this.DrawPolygon(polygon.Points[i])
		}
		this.dc.Clip()
		this.dc.InvertMask() // 反转mask是关键
		this.DrawPolygon(polygon.Points[0])
		this.dc.ResetClip() // 最后还要消除clip区域
	}
}

// 绘制简单多边形
func (this *Canvasgg) DrawPolygon(pnts []Point) {
	count := len(pnts)
	if count >= 3 {
		this.dc.MoveTo(float64(pnts[0].X), float64(pnts[0].Y))
		for i := 1; i < len(pnts)-1; i++ {
			this.dc.LineTo(float64(pnts[i].X), float64(pnts[i].Y))
		}
		this.dc.FillPreserve()
		this.dc.Stroke()
	}
}

// 绘制文字
func (this *Canvasgg) DrawString(text string, x, y float32) {
	// this.dc.DrawString(text, float64(x), float64(y))
	this.dc.DrawStringAnchored(text, float64(x), float64(y), 0.5, 0.5)
}

// =================================================== //

// // 和Canvasgg配合使用的坐标转化参数
// type coordParamsgg struct {
// 	Scale     float64      // 图片距离/地图距离
// 	MapCenter base.Point2D // 地图中心点
// 	// drawCenter Point        // 图片中心点
// 	dx, dy int // 图片的宽度和高度
// }

// // 根据地图的box和要绘制的大小，来初始化参数
// func (this *coordParamsgg) Init(bbox base.Rect2D, dx, dy int) {
// 	scaleX := (float64)(dx) / (bbox.Max.X - bbox.Min.X)
// 	scaleY := (float64)(dy) / (bbox.Max.Y - bbox.Min.Y)
// 	this.Scale = math.Min(scaleX, scaleY)

// 	this.MapCenter.X = (bbox.Max.X + bbox.Min.X) / 2
// 	this.MapCenter.Y = (bbox.Max.Y + bbox.Min.Y) / 2

// 	// this.drawCenter.X = dx / 2
// 	// this.drawCenter.Y = dy / 2
// 	this.dx = dx
// 	this.dy = dy
// 	// fmt.Println("params:", this)
// }

// // 得到当前地图的范围
// func (this *coordParamsgg) GetBounds() base.Rect2D {
// 	var bbox base.Rect2D
// 	dx := float64(this.dx) / 2.0 / this.Scale
// 	dy := float64(this.dy) / 2.0 / this.Scale
// 	bbox.Min.X = this.MapCenter.X - dx
// 	bbox.Min.Y = this.MapCenter.Y - dy
// 	bbox.Max.X = this.MapCenter.X + dx
// 	bbox.Max.Y = this.MapCenter.Y + dy
// 	// fmt.Println("CoordParams.GetBounds():", bbox)
// 	return bbox
// }

// // 正向转化一个点坐标：从地图坐标变为图片坐标
// // 绘制坐标 x = (pnt2D - mapCenter)*scale + drawCenter
// // 		    y = dy - ((pnt2D - mapCenter)*scale + drawCenter)
// func (this *coordParamsgg) Forward(pnt base.Point2D) Point {
// 	var drawPnt Point
// 	drawPnt.X = ((pnt.X - this.MapCenter.X) * this.Scale) + this.dx/2
// 	drawPnt.Y = this.dy - ((int)((pnt.Y-this.MapCenter.Y)*this.Scale) + this.dy/2)
// 	return drawPnt
// }

// func (this *coordParamsgg) Forwards(pnts []base.Point2D) []Point {
// 	drawPnts := make([]Point, len(pnts))
// 	for i, v := range pnts {
// 		drawPnts[i] = this.Forward(v)
// 	}
// 	return drawPnts
// }

// // 反向转化坐标
// // Reverse
