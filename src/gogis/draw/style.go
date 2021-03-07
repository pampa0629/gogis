package draw

import (
	"fmt"
	"image/color"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// 包初始化，这里设置随机数的不同
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 图层和对象绘制风格
type Style struct {
	LineColor Color // color.RGBA
	LineWidth float64
	LineDash  []float64
	FillColor Color // color.RGBA
}

// 定义一个color的目的是为了能输出 #RRGGBBAA 格式的json
type Color color.RGBA

// 颜色是否为空
func (this *Color) IsEmpty() bool {
	return this.A == 0
}

// 转化为浮点数，范围[0,1]
func (this *Color) ToFloat() (r, g, b, a float32) {
	r = float32(this.R) / 255.0
	g = float32(this.G) / 255.0
	b = float32(this.B) / 255.0
	a = float32(this.A) / 255.0
	return
}

func (this *Color) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`"#%02X%02X%02X%02X"`, this.R, this.G, this.B, this.A)
	return []byte(str), nil
}

// 十六进制字符串转 uint8
func strx2uint8(str string) uint8 {
	value, _ := strconv.ParseInt(str, 16, 0)
	return uint8(value)
}

func (this *Color) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "#\"")
	this.R = strx2uint8(str[0:2])
	this.G = strx2uint8(str[2:4])
	this.B = strx2uint8(str[4:6])
	this.A = strx2uint8(str[6:8])
	return nil
}

// 得到随机风格
func RandStyle() (style Style) {
	style.FillColor = Color(RandColor())
	style.LineColor = Color(RandColor())
	style.LineWidth = 1
	return
}

// 得到高亮风格
func HilightStyle() (style Style) {
	style.FillColor = Color(color.RGBA{0, 0, 255, 255})
	style.LineColor = Color(color.RGBA{0, 0, 255, 255})
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
