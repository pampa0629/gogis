package draw

import (
	"gogis/base"
	"image"
	"image/color"
)

type Point struct {
	X, Y float32
}

type Polyline struct {
	Points [][]Point
}

// 带洞的多边形（不带岛）
type Polygon struct {
	Points [][]Point
}

// 画布类型定义
type CanvasType string

const (
	Default CanvasType = "Default" //  默认用gg
	GG      CanvasType = "gg"
	GL      CanvasType = "gl"   // opengl
	GLSL    CanvasType = "glsl" // opengl shader language
)

// 抽象画布，具体可实现为：gg、gdi、opengl等不同类型
type Canvas interface {
	// 初始化，给出全幅的地理范围，以及画布的物理（屏幕）宽高;
	// data是附加参数，可根据不同具体画布来指定
	Init(bbox base.Rect2D, width, height int, data interface{})
	Clone() Canvas
	// 清空已经绘制的内容，清空后可继续绘制
	Clear()
	// 彻底删除画布，之后不可绘制
	Destroy()

	// 得到当前地图的范围
	GetBounds() base.Rect2D
	// 返回物理屏幕的尺寸
	GetSize() (int, int)

	// 正向转化一个点坐标：从地图坐标变为画布坐标
	Forward(pnt base.Point2D) Point
	Forwards(pnts []base.Point2D) []Point
	// 缩放画布中的物体; ratio为缩放比率，大于1为放大；小于1为缩小
	Zoom(ratio float64)
	Zoom2(ratio float64, x, y int)
	// 平移
	Pan(dx, dy int)
	// PanMap(dx, dy float32)

	// 得到go image对象，用来保存图片等操作
	GetImage() image.Image
	// 设置要绘制的风格 todo
	SetStyle(style Style)
	SetTextColor(textColor color.RGBA)

	DrawPoint(pnt Point)
	DrawLine(pnts []Point)
	DrawPolygon(pnts []Point)
	DrawPolyPolygon(polygon *Polygon)
	DrawString(text string, x, y float32)
	DrawImage(img image.Image, x, y float32)
}

// 是否支持在画布上绘制
type DrawCanvas interface {
	Draw(canvas Canvas)
}
