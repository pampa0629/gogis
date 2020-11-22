package draw

import (
	"image/color"
	"math/rand"
	"time"
)

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
	rand.Seed(time.Now().UnixNano())
	r := uint8(rand.Intn(255))
	g := uint8(rand.Intn(255))
	b := uint8(rand.Intn(255))
	return color.RGBA{r, g, b, 255}
}
