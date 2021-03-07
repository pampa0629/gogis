package draw

import (
	"crypto/md5"
	"fmt"
	"gogis/base"
	"image"
	"image/color"
	"strings"

	// "github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/gl/v4.1-core/gl"
)

// 适合opengl sl的画布
type Canvassl struct {
	CoordParams
	style         Style
	width, height int
	d2v           map[string]uint32

	ratio float64
	// xoff, yoff float64
	offset base.Point2D

	prog uint32
}

func (this *Canvassl) Clone() Canvas {
	var canvas = new(Canvassl)
	canvas.CoordParams = this.CoordParams
	canvas.SetStyle(this.style)
	// todo d2v
	return canvas
}

// 初始化: 计算坐标转化参数，构造dc
func (this *Canvassl) Init(bbox base.Rect2D, width, height int, data interface{}) {
	this.CoordParams.Init(bbox, 2, 2, true, Point{1, 1})
	this.width, this.height = width, height
	this.d2v = make(map[string]uint32)
	this.ratio = 1
	this.linkShader()
	// this.flushRatio()
}

// func (this *Canvassl) flushRatio() {
// 	r := "ratio"
// 	rdata := []byte(r)
// 	loc := gl.GetUniformLocation(this.prog, (*uint8)(&rdata[0]))
// 	gl.ProgramUniform1f(this.prog, loc, float32(this.ratio))
// }

func (this *Canvassl) Zoom(ratio float64) {
	// return this.width, this.height
	this.ratio *= ratio // 累加
	this.linkShader()
	// this.flushRatio()
}

// todo
func (this *Canvassl) Zoom2(ratio float64, x, y int) {
	this.ratio *= ratio // 累加

	// 先归一化到[0,2]
	rx := float64(x) / float64(this.width) * 2.0
	ry := 2.0 - float64(y)/float64(this.height)*2.0

	// 计算中心点偏移量
	dx := (ratio - 1) + rx*(1-ratio)
	dy := (ratio - 1) + ry*(1-ratio)

	this.offset.X += dx / this.Scale * this.ratio
	this.offset.Y -= dy / this.Scale * this.ratio

	this.linkShader()
}

// todo
func (this *Canvassl) Pan(dx, dy int) {
	// 先归一化到[0,2]
	rx := float64(dx) / float64(this.width) * 2.0
	ry := float64(dy) / float64(this.height) * 2.0
	this.offset.X += rx
	this.offset.Y -= ry
	this.linkShader()
}

func (this *Canvassl) GetBounds() (bbox base.Rect2D) {
	bbox = this.CoordParams.GetBounds()
	bbox = bbox.Scale(1.0 / this.ratio)
	offx := -this.offset.X / this.Scale
	offy := -this.offset.Y / this.Scale
	bbox = bbox.Offset(offx, offy)
	return
}

// 清空DC，为下次绘制做好准备
func (this *Canvassl) Clear() {
}

// todo
func (this *Canvassl) Destroy() {
	this.Clear()
	this.d2v = make(map[string]uint32)
}

// todo
func (this *Canvassl) GetImage() image.Image {
	// return this.dc.Image()
	return nil
}

func (this *Canvassl) GetSize() (width, height int) {
	return this.width, this.height
}

func (this *Canvassl) SetStyle(style Style) {
	this.style = style
	gl.LineWidth(float32(style.LineWidth * 2))
	if !style.LineColor.IsEmpty() {
		// r, g, b, _ := style.LineColor.ToFloat()
		// gl.Color3f(r, g, b)
	}
	if !style.FillColor.IsEmpty() {
		// r, g, b, _ := style.FillColor.ToFloat()
		// gl.Color3f(r, g, b)
	}
	this.linkShader()
}

// todo
func (this *Canvassl) SetTextColor(textColor color.RGBA) {
	// pat := gg.NewSolidPattern(textColor)
	// this.dc.SetFillStyle(pat)
}

// todo
func (this *Canvassl) DrawImage(img image.Image, x, y float32) {
	// this.dc.DrawImage(img, x, y)
}

func (this *Canvassl) DrawPoint(pnt Point) {
	// this.dc.DrawPoint(float64(pnt.X), float64(pnt.Y), 1)
	// this.dc.Stroke()
	// x, y := this.ChangePoint(pnt)
	// gl.Begin(gl.POINTS)
	// gl.Vertex3f(pnt.X, pnt.Y, 0.0)
	// gl.End()
}

func (this *Canvassl) drawLineVao(vao uint32, count int32) {
	gl.BindVertexArray(vao)
	// gl.DrawArrays(gl.TRIANGLES, 0, int32(len(polygon41)/3))
	gl.DrawArrays(gl.LINE_STRIP, 0, count)
}

func (this *Canvassl) DrawLine(pnts []Point) {
	sum := md5.Sum(base.ByteSlice(pnts))
	md := string(sum[:])
	if vao, ok := this.d2v[md]; ok {
		this.drawLineVao(vao, int32(len(pnts)))
	} else {
		vao := makePntsVao(pnts)
		this.d2v[md] = vao
		this.drawLineVao(vao, int32(len(pnts)))
	}
}

// 绘制复杂面（带洞）
// len必须大于1；[0] 是面，后面的都是洞；点的绕圈方向不论
func (this *Canvassl) DrawPolyPolygon(polygon *Polygon) {
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

// todo
// 绘制简单多边形
func (this *Canvassl) DrawPolygon(pnts []Point) {
}

// todo
// 绘制文字
func (this *Canvassl) DrawString(text string, x, y float32) {
	// this.dc.DrawStringAnchored(text, float64(x), float64(y), 0.5, 0.5)
}

func makePntsVao(points []Point) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 8*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, nil)
	return vao
}

// ================================== //
const (
	vertexShaderSource = `
    #version 410
    layout (location = 0) in vec3 vp;
	void main() {
		float x = vp.x * %v + %v;
		float y = vp.y * %v + %v;
        gl_Position = vec4(x,y,vp.z, 1.0);
    }`
	vertexShaderSource2 = `
    #version 410
    layout (location = 0) in vec3 vp;
	uniform float ratio; // why not?
    void main() {
        gl_Position = vec4(vp*ratio, 1.0);
    }`
	fragmentShaderSource = `
    #version 410
    layout (location = 0) out vec4 frag_colour;
    void main() {
        frag_colour = vec4(%v, %v, %v, %v);
    }`
)

func (this *Canvassl) buildVS() string {
	// offx := this.offset.X / this.CoordParams.Scale
	// offy := this.offset.Y / this.CoordParams.Scale
	vertexShaderSrc := fmt.Sprintf(vertexShaderSource, this.ratio, this.offset.X, this.ratio, this.offset.Y)
	vertexShaderSrc += "\x00"
	return vertexShaderSrc
}

func (this *Canvassl) linkShader() {
	// tr := base.NewTimeRecorder()

	// vertexShaderSrc := vertexShaderSource2 + "\x00"
	vertexShaderSrc := this.buildVS()
	vertexShader, err := compileShader(vertexShaderSrc, gl.VERTEX_SHADER)
	base.PrintError(vertexShaderSrc, err)
	if err != nil {
		// fmt.Println("vertex shader error:", err)
		panic(err)
	}

	r, g, b, a := this.style.LineColor.ToFloat()
	fragmentShaderSrc := fmt.Sprintf(fragmentShaderSource, r, g, b, a)
	fragmentShaderSrc += "\x00"
	fragmentShader, err := compileShader(fragmentShaderSrc, gl.FRAGMENT_SHADER)
	base.PrintError(fragmentShaderSrc, err)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	gl.UseProgram(prog)
	this.prog = prog
	// tr.Output("prog")
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}
	return shader, nil
}
