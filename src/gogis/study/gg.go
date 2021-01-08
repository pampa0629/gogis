package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/fogleman/gg"
)

func main() {
	ggText()
	// ggDong()
	fmt.Println("gg DONE")
}

func ggText() {
	dc := gg.NewContext(1000, 1000)
	pat := gg.NewSolidPattern(color.RGBA{25, 200, 20, 255})
	dc.SetFillStyle(pat)
	// dc.set

	dc.DrawString("abc", 100, 100)
	dc.DrawPoint(100, 100, 1)

	dc.DrawStringAnchored("abc", 300, 300, 0.5, 0.5)
	dc.DrawPoint(300, 300, 1)

	// 多行文本
	dc.DrawStringWrapped("abc", 500, 500, 0, 0, 500, 0, gg.AlignLeft)
	dc.DrawPoint(500, 500, 1)

	dc.Fill()

	dc.SavePNG("c:/temp/gg.png")
}

func ggColor() {
	dc1 := gg.NewContext(1000, 1000)

	// pat := gg.NewSolidPattern(color.RGBA{25, 200, 20, 255})
	// dc1.SetFillStyle(pat)
	dc1.SetFillColor(color.RGBA{25, 200, 20, 255})
	dc1.SetStrokeColor(color.RGBA{255, 20, 20, 255})

	dc1.MoveTo(200, 200)
	dc1.LineTo(200, 400)
	dc1.LineTo(400, 400)
	dc1.LineTo(400, 200)
	dc1.LineTo(200, 200)
	dc1.FillPreserve()
	dc1.Stroke()

	// dc1.MoveTo(200, 200)
	// dc1.LineTo(200, 400)
	// dc1.LineTo(400, 400)
	// dc1.LineTo(400, 200)
	// dc1.LineTo(200, 200)
	// dc1.Stroke()

	dc1.SavePNG("c:/temp/out1.png")
}

// 岛洞
func ggDong() {
	dc1 := gg.NewContext(1000, 1000)
	dc1.Clear()
	pat := gg.NewSolidPattern(color.RGBA{25, 200, 20, 255})
	dc1.SetFillStyle(pat)

	dc1.MoveTo(200, 200)
	dc1.LineTo(200, 400)
	dc1.LineTo(400, 400)
	dc1.LineTo(400, 200)
	dc1.LineTo(200, 200)

	dc1.MoveTo(600, 200)
	dc1.LineTo(800, 200)
	dc1.LineTo(800, 400)
	dc1.LineTo(600, 400)
	dc1.LineTo(600, 200)
	// dc1.Fill()
	dc1.Clip()
	dc1.InvertMask()

	dc1.MoveTo(100, 100)
	dc1.LineTo(100, 900)
	dc1.LineTo(900, 900)
	dc1.LineTo(900, 100)
	dc1.LineTo(100, 100)
	dc1.Fill()

	dc1.ResetClip()

	dc1.MoveTo(650, 250)
	dc1.LineTo(750, 250)
	dc1.LineTo(750, 350)
	dc1.LineTo(650, 350)
	dc1.LineTo(650, 250)
	dc1.Fill()

	dc1.SavePNG("c:/temp/out1.png")
}

func ggFill() {
	dc1 := gg.NewContext(1000, 1000)
	dc1.Clear()
	pat := gg.NewSolidPattern(color.RGBA{255, 1, 0, 255})
	dc1.SetFillStyle(pat)
	dc1.MoveTo(0, 0)
	dc1.LineTo(500, 0)
	dc1.LineTo(500, 500)
	dc1.LineTo(0, 500)
	dc1.LineTo(0, 0)
	dc1.Fill()

	dc1.SavePNG("c:/temp/out1.png")
}

func ggMask() {
	dc1 := gg.NewContext(1000, 1000)
	dc1.Clear()
	pat := gg.NewSolidPattern(color.RGBA{255, 1, 0, 255})
	dc1.SetFillStyle(pat)
	dc1.DrawCircle(400, 400, 400)
	dc1.Fill()
	dc1.SavePNG("c:/temp/out1.png")

	dc2 := gg.NewContext(1000, 1000)
	pat2 := gg.NewSolidPattern(color.RGBA{0, 255, 0, 100})
	dc2.SetFillStyle(pat2)
	dc2.DrawCircle(400, 600, 400)
	dc2.Fill()
	dc2.SavePNG("c:/temp/out2.png")

	dc3 := gg.NewContext(1000, 1000)
	dc3.DrawImage(dc1.Image(), 0, 0)
	dc3.DrawImage(dc2.Image(), 0, 0)
	dc3.SavePNG("c:/temp/out3.png")

}

func ggImage() {
	im, err := gg.LoadImage("c:/temp/baboon.png")
	if err != nil {
		log.Fatal(err)
	}

	dc := gg.NewContext(512, 512)
	dc.DrawCircle(256, 256, 256)
	dc.Clip()
	dc.DrawImage(im, 0, 0)
	dc.SavePNG("c:/temp/out.png")
}

func ggImageAlpha() {
	dc := gg.NewContext(1000, 1000)
	// dc.SetRGB(255, 0, 0)
	pat := gg.NewSolidPattern(color.RGBA{255, 1, 0, 255})
	dc.SetFillStyle(pat)
	dc.DrawCircle(400, 400, 400)
	dc.Fill()
	// dc.Stroke() //没有这句是不会把线最终画出来的
	err := dc.SavePNG("c:/temp/out.png")
	fmt.Println("err:", err)
}

func ggtest2() {
	dc := gg.NewContext(1000, 1000)
	dc.DrawCircle(500, 500, 400)
	dc.SetRGB(255, 0, 0)
	dc.Fill()
	err := dc.SavePNG("c:/temp/out.png")
	fmt.Println("err:", err)
}

func ggtest() {

	const W = 1024
	const H = 1024
	dc := gg.NewContext(W, H) //上下文，含长和宽
	dc.SetRGB(0, 0, 0)        //设置当前色
	dc.Clear()                //清理一下上下文，下面开始画画

	for i := 0; i < 10; i++ { //画1000 条线，随机位置，长度，颜色和透明度
		x1 := rand.Float64() * W
		y1 := rand.Float64() * H
		x2 := rand.Float64() * W
		y2 := rand.Float64() * H

		r := rand.Float64()
		g := rand.Float64()
		b := rand.Float64()
		a := rand.Float64()*0.5 + 0.5
		w := rand.Float64()*4 + 5
		dc.SetRGBA(r, g, b, a)
		dc.SetDash(20, 20)
		dc.SetLineCap(gg.LineCapRound)
		dc.SetLineWidth(w)
		dc.DrawLine(x1, y1, x2, y2) //画线
		dc.DrawEllipse(200, 200, 300, 400)
		pat := gg.NewSolidPattern(color.RGBA{233, 1, 1, 233})
		dc.SetFillStyle(pat)
		dc.Fill()
		dc.Stroke() //没有这句是不会把线最终画出来的
	}

	dc.DrawString("hello", 100, 100)
	dc.SavePNG("c:/temp/lines.png") //保存上下文为一张图片
	fmt.Println("Done")

}
