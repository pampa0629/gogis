package draw

import (
	"image/color"
	"math/rand"
	"time"
)

// 包初始化，这里设置随机数的不同
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 图层和对象绘制风格
type Style struct {
	LineColor color.RGBA
	LineWidth float64
	LineDash  []float64
	FillColor color.RGBA
}

// 得到随机风格
func RandStyle() (style Style) {
	style.FillColor = RandColor()
	style.LineColor = RandColor()
	style.LineWidth = 1
	return
}

// 得到随机颜色
func RandColor() color.RGBA {
	r := uint8(rand.Intn(200))
	g := uint8(rand.Intn(200))
	b := uint8(rand.Intn(200))
	return color.RGBA{r, g, b, 255}
}
