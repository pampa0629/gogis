package desktop

import (
	"fmt"
	"gogis/base"
	"math"

	_ "gogis/data/memory"
	_ "gogis/data/shape"
	_ "gogis/data/sqlite"
	"gogis/mapping"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	vertexShaderSource = `
    #version 410
    layout (location = 0) in vec3 vp;
    void main() {
        gl_Position = vec4(vp, 1.0);
    }
` + "\x00"
	fragmentShaderSource = `
    #version 410
    layout (location = 0) out vec4 frag_colour;
    void main() {
        frag_colour = vec4(1, 0, 0, 1);
    }
` + "\x00"
)

// var (
// 	width  = 1024
// 	height = 768
// )

func KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	fmt.Println("key:", key, scancode, action, mods)
}

var mousex, mousey float64

func MouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	fmt.Println("mouse:", button, action, mods)
	if action == glfw.Press {
		mousex, mousey = w.GetCursorPos()
	} else if action == glfw.Release {
		x, y := w.GetCursorPos()
		glmap.Pan(int(x-mousex), int(y-mousey))
	}

	// fmt.Println("cursor:", x, y)
}

func ScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	fmt.Println("scroll:", xoff, yoff)
	ratio := math.Pow(1.5, yoff)
	glmap.Zoom(ratio)
	// x, y := w.GetCursorPos()
	// glmap.Zoom2(ratio, int(x), int(y))

}

var glmap *mapping.Map

func ShowGL(gmap *mapping.Map, width, height int) {
	glmap = gmap
	runtime.LockOSThread()
	window := initGlfw(width, height)
	defer glfw.Terminate()

	initOpenGL()
	gmap.Prepare(width, height)

	window.SetKeyCallback(KeyCallback)
	window.SetMouseButtonCallback(MouseButtonCallback)
	window.SetScrollCallback(ScrollCallback)

	// linkShader()
	for !window.ShouldClose() {
		// tr := base.NewTimeRecorder()
		drawgl(window, gmap)
		// tr.Output("draw")
	}
}

func drawgl(window *glfw.Window, gmap *mapping.Map) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gmap.Draw()
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
func initOpenGL() {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)
}

func linkShader() {
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
	gl.UseProgram(prog)
}

// initGlfw 初始化 glfw 并且返回一个可用的窗口。
func initGlfw(width, height int) *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(width, height, "gogis opengl", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	// window.GetWin32Window()
	return window
}
