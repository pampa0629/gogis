package draw

import (
	"gogis/base"
	"image"
	"image/color"

	// "github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/gl/v2.1/gl"
)

// 适合opengl的画布；注意，采用的是opengl绘制，而非opengl sl的绘制
type Canvasgl struct {
	CoordParams
	style         Style
	width, height int
}

func (this *Canvasgl) Clone() Canvas {
	var canvas = new(Canvasgl)
	canvas.CoordParams = this.CoordParams
	canvas.SetStyle(this.style)
	return canvas
}

// 初始化: 计算坐标转化参数，构造dc
func (this *Canvasgl) Init(bbox base.Rect2D, width, height int, data interface{}) {
	this.CoordParams.Init(bbox, 2, 2, true, Point{1, 1})
	this.width, this.height = width, height
}

// 清空DC，为下次绘制做好准备
// todo
func (this *Canvasgl) Clear() {
	// this.dc.Clear()
}

// todo
func (this *Canvasgl) Destroy() {

}

// todo
func (this *Canvasgl) GetImage() image.Image {
	// return this.dc.Image()
	return nil
}

func (this *Canvasgl) GetSize() (int, int) {
	return this.width, this.height
}

// 检查是否真正在image中绘制了
func (this *Canvasgl) CheckDrawn() bool {
	// img := this.dc.Image()
	// pix := img.(*image.RGBA).Pix
	// if pix != nil {
	// 	for _, v := range pix {
	// 		if v != 0 {
	// 			return true
	// 		}
	// 	}
	// }
	// todo
	return false
}

func (this *Canvasgl) SetStyle(style Style) {
	this.style = style
	gl.LineWidth(float32(style.LineWidth * 2))
	if !style.LineColor.IsEmpty() {
		r, g, b, _ := style.LineColor.ToFloat()
		gl.Color3f(r, g, b)
	}
	if !style.FillColor.IsEmpty() {
		r, g, b, _ := style.FillColor.ToFloat()
		gl.Color3f(r, g, b)
	}
}

// todo
func (this *Canvasgl) SetTextColor(textColor color.RGBA) {
	// pat := gg.NewSolidPattern(textColor)
	// this.dc.SetFillStyle(pat)
}

// todo
func (this *Canvasgl) DrawImage(img image.Image, x, y float32) {
	// this.dc.DrawImage(img, x, y)
}

func (this *Canvasgl) DrawPoint(pnt Point) {
	gl.Begin(gl.POINTS)
	gl.Vertex3f(pnt.X, pnt.Y, 0.0)
	gl.End()
}

func (this *Canvasgl) DrawLine(pnts []Point) {
	gl.Begin(gl.LINE_STRIP)
	for i := 0; i < len(pnts); i++ {
		gl.Vertex3f(pnts[i].X, pnts[i].Y, 0)
	}
	gl.End()
}

// 绘制复杂线
// func (this *Canvasgl) DrawPolyPolyline(polyline *Polyline) {
// 	for _, v := range polyline.Points {
// 		this.DrawPolyline(v)
// 	}
// }

// todo 应该用glu.tess来解决岛洞的绘制
// 绘制复杂面（带洞）
// len必须大于1；[0] 是面，后面的都是洞；点的绕圈方向不论
func (this *Canvasgl) DrawPolyPolygon(polygon *Polygon) {
	polyCount := len(polygon.Points)
	if polyCount == 1 {
		// 简单多边形
		this.DrawPolygon(polygon.Points[0])
	} else if polyCount > 1 {
		// 先绘制后面的洞，再clip、mask一下，最后绘制面
		for i := 1; i < polyCount; i++ {
			this.DrawPolygon(polygon.Points[i])
		}
		// this.dc.Clip()
		// this.dc.InvertMask() // 反转mask是关键
		this.DrawPolygon(polygon.Points[0])
		// this.dc.ResetClip() // 最后还要消除clip区域
	}
}

// 绘制简单多边形
func (this *Canvasgl) DrawPolygon(pnts []Point) {
	gl.Begin(gl.POLYGON)
	for i := 0; i < len(pnts); i++ {
		gl.Vertex3f(pnts[i].X, pnts[i].Y, 0)
	}
	gl.End()
}

// todo
// 绘制文字
func (this *Canvasgl) DrawString(text string, x, y float32) {
	// this.dc.DrawStringAnchored(text, float64(x), float64(y), 0.5, 0.5)
}

// ======================================================== //

// // 按照opengl的要求，坐标转化到[-1,1]之间的坐标转化参数
// type coordParamsgl struct {
// 	Scale      float64      // 图片距离/地图距离
// 	MapCenter  base.Point2D // 地图中心点
// 	drawCenter Point        // 图片中心点
// 	dx, dy     int          // 图片的宽度和高度
// }

// // 根据地图的box和要绘制的大小，来初始化参数
// func (this *coordParamsgl) Init(bbox base.Rect2D, dx int, dy int) {
// 	scaleX := 2.0 / bbox.Dx()
// 	scaleY := 2.0 / bbox.Dy()
// 	this.Scale = math.Min(scaleX, scaleY)

// 	this.MapCenter = bbox.Center()

// 	// this.drawCenter.X = float32(dx) / 2.0
// 	// this.drawCenter.Y = float32(dy) / 2.0
// 	this.dx = dx
// 	this.dy = dy
// 	// fmt.Println("params:", this)
// }

// // 得到当前地图的范围
// func (this *coordParamsgl) GetBounds() base.Rect2D {
// 	var bbox base.Rect2D
// 	delta := 1.0 / this.Scale
// 	bbox.Min.X = this.MapCenter.X - delta
// 	bbox.Min.Y = this.MapCenter.Y - delta
// 	bbox.Max.X = this.MapCenter.X + delta
// 	bbox.Max.Y = this.MapCenter.Y + delta
// 	// fmt.Println("CoordParams.GetBounds():", bbox)
// 	return bbox
// }

// // 正向转化一个点坐标：从地图坐标变为图片坐标
// // 绘制坐标 x = (pnt2D - mapCenter)*scale + drawCenter
// // 		    y = dy - ((pnt2D - mapCenter)*scale + drawCenter)
// func (this *coordParamsgl) Forward(pnt base.Point2D) Point {
// 	var drawPnt Point
// 	drawPnt.X = (float32)((pnt.X - this.MapCenter.X) * this.Scale)
// 	drawPnt.Y = (float32)((pnt.Y - this.MapCenter.Y) * this.Scale)
// 	// drawPnt.Y = this.dy - ((int)((pnt.Y-this.MapCenter.Y)*this.Scale) + this.drawCenter.Y)
// 	return drawPnt
// }

// func (this *coordParamsgl) Forwards(pnts []base.Point2D) []Point {
// 	drawPnts := make([]Point, len(pnts))
// 	for i, v := range pnts {
// 		drawPnts[i] = this.Forward(v)
// 	}
// 	return drawPnts
// }

// func (this *coordParamsgl) Forward32s(pnts []base.Point2D) []float32 {
// 	drawPnts := make([]float32, len(pnts)*2)
// 	for i, v := range pnts {
// 		fwd := this.Forward(v)
// 		drawPnts[2*i] = fwd.X
// 		drawPnts[2*i+1] = fwd.Y
// 	}
// 	return drawPnts
// }

// // 反向转化坐标
// // Reverse
