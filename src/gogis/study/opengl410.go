package main

// import (
// 	"fmt"
// 	"gogis/base"
// 	"runtime"
// 	"strings"

// 	// OR: github.com/go-gl/gl/v2.1/gl
// 	"github.com/go-gl/gl/v4.1-core/gl"

// 	"github.com/go-gl/glfw/v3.3/glfw"
// )

// const (
// 	width              = 1024
// 	height             = 768
// 	vertexShaderSource = `
//     #version 410
//     in vec3 vp;
//     void main() {
//         gl_Position = vec4(vp, 1.0);
//     }
// ` + "\x00"
// 	fragmentShaderSource = `
//     #version 410
//     out vec4 frag_colour;
//     void main() {
//         frag_colour = vec4(1, 0, 0, 1);
//     }
// ` + "\x00"
// )

// func main() {
// 	runtime.LockOSThread()
// 	window := initGlfw()
// 	defer glfw.Terminate()

// 	program := initOpenGL()
// 	// vao := makeVao(triangle)
// 	// vao := makeVao(polygon41)

// 	for !window.ShouldClose() {
// 		tr := base.NewTimeRecorder()
// 		drawgl(vao, window, program)
// 		tr.Output("draw")
// 	}
// }

// var (
// 	triangle = []float32{
// 		0, 0.1, 0, // top
// 		0.1, -0.1, 0, // left
// 		-0.1, -0.1, 0, // right
// 	}

// 	polygon41 = []float32{
// 		-0.05, 0, 0,
// 		-0.1, 0.1, 0,
// 		0.1, 0.1, 0,
// 		0.05, 0, 0,
// 		0.1, -0.1, 0,
// 		-0.1, -0.1, 0,
// 		-0.05, 0, 0}

// 	poly2gon41 = []float32{
// 		-0.05, 0,
// 		-0.1, 0.1,
// 		0.1, 0.1,
// 		0.05, 0,
// 		0.1, -0.1,
// 		-0.1, -0.1,
// 		-0.05, 0}
// )

// var vao uint32

// // makeVao 执行初始化并从提供的点里面返回一个顶点数组
// func makeVao(points []float32) uint32 {
// 	var vbo uint32
// 	gl.GenBuffers(1, &vbo)
// 	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
// 	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(&points[0]), gl.STATIC_DRAW)

// 	gl.GenVertexArrays(1, &vao)
// 	gl.BindVertexArray(vao)
// 	gl.EnableVertexAttribArray(0)
// 	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
// 	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, nil)
// 	return vao
// }

// func drawgl(vao uint32, window *glfw.Window, program uint32) {
// 	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
// 	gl.UseProgram(program)

// 	if vao == 0 {
// 		vao = makeVao(poly2gon41)
// 	}

// 	for i := 0; i < 100000; i++ {
// 		gl.BindVertexArray(vao)
// 		// gl.DrawArrays(gl.TRIANGLES, 0, int32(len(polygon41)/3))
// 		gl.DrawArrays(gl.LINE_STRIP, 0, int32(len(poly2gon41)/2))
// 	}

// 	glfw.PollEvents()
// 	window.SwapBuffers()
// }

// func compileShader(source string, shaderType uint32) (uint32, error) {
// 	shader := gl.CreateShader(shaderType)
// 	csources, free := gl.Strs(source)
// 	gl.ShaderSource(shader, 1, csources, nil)
// 	free()
// 	gl.CompileShader(shader)
// 	var status int32
// 	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
// 	if status == gl.FALSE {
// 		var logLength int32
// 		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
// 		log := strings.Repeat("\x00", int(logLength+1))
// 		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
// 		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
// 	}
// 	return shader, nil
// }

// // initOpenGL 初始化 OpenGL 并且返回一个初始化了的程序。
// func initOpenGL() uint32 {
// 	if err := gl.Init(); err != nil {
// 		panic(err)
// 	}
// 	version := gl.GoStr(gl.GetString(gl.VERSION))
// 	fmt.Println("OpenGL version", version)

// 	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
// 	if err != nil {
// 		panic(err)
// 	}

// 	prog := gl.CreateProgram()
// 	gl.AttachShader(prog, vertexShader)
// 	gl.AttachShader(prog, fragmentShader)
// 	gl.LinkProgram(prog)
// 	return prog
// }

// // initGlfw 初始化 glfw 并且返回一个可用的窗口。
// func initGlfw() *glfw.Window {
// 	if err := glfw.Init(); err != nil {
// 		panic(err)
// 	}
// 	glfw.WindowHint(glfw.Resizable, glfw.True)
// 	glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
// 	glfw.WindowHint(glfw.ContextVersionMinor, 1)
// 	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
// 	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
// 	window, err := glfw.CreateWindow(width, height, "Conway's Game of Life", nil, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	window.MakeContextCurrent()
// 	// window.GetWin32Window()
// 	return window
// }
