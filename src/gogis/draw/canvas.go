package draw

import (
	"gogis/base"
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

type Point image.Point

type Polyline struct {
	Points [][]Point
}

// 带洞的多边形（不带岛）
type Polygon struct {
	Points [][]Point
}

// 画布
type Canvas struct {
	dc     *gg.Context
	Params CoordParams
	style  Style
}

func (this *Canvas) Clone() *Canvas {
	var canvas = new(Canvas)
	canvas.Params = this.Params
	canvas.dc = gg.NewContext(canvas.Params.dx, canvas.Params.dy)
	canvas.SetStyle(this.style)
	return canvas
}

// 初始化: 计算坐标转化参数，构造dc
func (this *Canvas) Init(bbox base.Rect2D, width, height int) {
	this.Params.Init(bbox, width, height)
	this.dc = gg.NewContext(width, height)
	// fmt.Println("dc:", this.dc)
}

// 清空DC，为下次绘制做好准备
func (this *Canvas) ClearDC() {
	this.dc.Clear()
}

func (this *Canvas) Image() image.Image {
	return this.dc.Image()
}

func (this *Canvas) GetSize() (width, height int) {
	return this.dc.Width(), this.dc.Height()
}

// 检查是否真正在image中绘制了
func (this *Canvas) CheckDrawn() bool {
	img := this.dc.Image()
	pix := img.(*image.RGBA).Pix
	if pix != nil {
		for _, v := range pix {
			if v != 0 {
				return true
			}
		}
	}
	return false
}

func (this *Canvas) SetStyle(style Style) {
	this.style = style
	// this.dc.SetFillStyle(gg.NewSolidPattern(style.FillColor))
	this.dc.SetFillColor(style.FillColor)
	this.dc.SetStrokeColor(style.LineColor)

	// this.dc.SetColor(style.LineColor)
	this.dc.SetLineWidth(style.LineWidth)
	this.dc.SetDash(style.LineDash...)
}

func (this *Canvas) SetTextColor(textColor color.RGBA) {
	pat := gg.NewSolidPattern(textColor)
	this.dc.SetFillStyle(pat)
}

func (this *Canvas) Stroke() {
	this.dc.Stroke()
}

func (this *Canvas) DrawImage(img image.Image, x, y int) {
	this.dc.DrawImage(img, x, y)
}

func (this *Canvas) DrawPoint(pnt Point) {
	this.dc.DrawPoint(float64(pnt.X), float64(pnt.Y), 1)
	this.dc.Stroke()
	// 先画个小方框代表点
	// var pnts [5]Point
	// pnts[0].X = pnt.X - 1
	// pnts[0].Y = pnt.Y - 1
	// pnts[1].X = pnt.X + 1
	// pnts[1].Y = pnt.Y - 1
	// pnts[2].X = pnt.X + 1
	// pnts[2].Y = pnt.Y + 1
	// pnts[3].X = pnt.X - 1
	// pnts[3].Y = pnt.Y + 1
	// pnts[4].X = pnt.X - 1
	// pnts[4].Y = pnt.Y - 1
	// this.DrawPolyline(pnts[:])
}

func (this *Canvas) DrawPolyline(pnts []Point) {
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
func (this *Canvas) DrawPolyPolyline(polyline *Polyline) {
	for _, v := range polyline.Points {
		this.DrawPolyline(v)
	}
}

// 绘制复杂面（带洞）
// len必须大于1；[0] 是面，后面的都是洞；点的绕圈方向不论
func (this *Canvas) DrawPolyPolygon(polygon *Polygon) {
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
func (this *Canvas) DrawPolygon(pnts []Point) {
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
func (this *Canvas) DrawString(text string, x, y int) {
	// this.dc.DrawString(text, float64(x), float64(y))
	this.dc.DrawStringAnchored(text, float64(x), float64(y), 0.5, 0.5)
}

// 是否支持在画布上绘制
type DrawCanvas interface {
	Draw(canvas *Canvas)
}
