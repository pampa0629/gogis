package main

import (
	"fmt"
	"gogis/base"
	"gogis/draw"
	"gogis/mapping"
	"runtime"
	"strings"

	// OR: github.com/go-gl/gl/v2.1/gl
	"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	width              = 1600
	height             = 1200
	vertexShaderSource = `
    #version 410
    in vec3 vp;
    void main() {
        gl_Position = vec4(vp, 1.0);
    }
` + "\x00"
	fragmentShaderSource = `
    #version 410
    out vec4 frag_colour;
    void main() {
        frag_colour = vec4(1, 0, 0, 1);
    }
` + "\x00"
)

func initMap412() *mapping.Map {
	gmap := mapping.NewMap(draw.GLSL)
	gmap.Open("c:/temp/dltb_line.gmp") // railway chinapnt_84 County_R Provinces JBNTBHTB  dltb
	gmap.Prepare(width, height)
	// gmap.Zoom(2.1)
	return gmap
}

func main() {
	runtime.LockOSThread()
	window := initGlfw()
	defer glfw.Terminate()

	program := initOpenGL()
	gmap := initMap412()
	// vao := makeVao(triangle)
	vao := makeVao(polygon41)
	// gl.MatrixMode(gl.PROJECTION)
	// glu.Ortho2D(-100, 100, -100, 100)

	for !window.ShouldClose() {
		tr := base.NewTimeRecorder()
		drawgl412(window, program, gmap, vao)
		tr.Output("draw")
	}
}

var (
	triangle = []float32{
		0, 0.1, 0, // top
		0.1, -0.1, 0, // left
		-0.1, -0.1, 0, // right
	}

	polygon41 = []float32{
		-5, 0, 0,
		-10, 10, 0,
		10, 10, 0,
		5, 0, 0,
		10, -10, 0,
		-10, -10, 0,
		-5, 0, 0}
)

// makeVao 执行初始化并从提供的点里面返回一个顶点数组
func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(&points[0]), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
	return vao
}

func drawgl412(window *glfw.Window, program uint32, gmap *mapping.Map, vao uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gmap.Draw()
	// for i := 0; i < 100000; i++ {
	// 	gl.BindVertexArray(vao)
	// 	// 	// gl.DrawArrays(gl.TRIANGLES, 0, int32(len(polygon41)/3))
	// 	gl.DrawArrays(gl.LINE_STRIP, 0, int32(len(polygon41)/3))
	// }

	glfw.PollEvents()
	window.SwapBuffers()
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

// initOpenGL 初始化 OpenGL 并且返回一个初始化了的程序。
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	tr := base.NewTimeRecorder()
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	tr.Output("prog")
	return prog
}

// initGlfw 初始化 glfw 并且返回一个可用的窗口。
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(width, height, "Conway's Game of Life", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	// window.GetWin32Window()
	return window
}
