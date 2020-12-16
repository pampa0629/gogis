package draw

import (
	"testing"
)

// 矩形与点
func TestStyle(t *testing.T) {
	{
		c1 := RandColor()
		c2 := RandColor()
		if c1 == c2 {
			t.Errorf("随机风格错误1")
		}
	}
	{
		style := RandStyle()
		if style.FillColor == style.LineColor {
			t.Errorf("随机风格错误2")
		}
	}
}
