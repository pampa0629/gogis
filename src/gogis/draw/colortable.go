package draw

import (
	"image/color"
)

// 颜色表
type ColorTable struct {
	// colors []color.RGBA
	Max, Min color.RGBA
}

func (this *ColorTable) Make(max, min color.RGBA) {
	this.Max = max
	this.Min = min
}

func (this *ColorTable) GetColor(value, max, min int) color.RGBA {
	ratio := float64(value-min) / float64(max-min)
	r := uint8(float64(this.Max.R-this.Min.R)*ratio) + this.Min.R
	g := uint8(float64(this.Max.G-this.Min.G)*ratio) + this.Min.G
	b := uint8(float64(this.Max.B-this.Min.B)*ratio) + this.Min.B
	a := uint8(float64(this.Max.A-this.Min.A)*ratio) + this.Min.A
	return color.RGBA{r, g, b, a}
}
