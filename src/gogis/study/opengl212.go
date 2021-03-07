package main

// import (
// 	"fmt"
// 	"gogis/base"
// 	_ "gogis/data/shape"
// 	_ "gogis/data/sqlite"
// 	"gogis/draw"
// 	"gogis/mapping"
// 	"runtime"
// 	"unsafe"

// 	// OR: github.com/go-gl/gl/v2.1/gl
// 	// "github.com/go-gl/gl" "github.com/go-gl-legacy/glu"
// 	"github.com/go-gl-legacy/glu"
// 	"github.com/go-gl/gl/v2.1/gl"

// 	"github.com/go-gl/glfw/v3.2/glfw"
// )

// // const (
// // 	width  = 1024
// // 	height = 768
// // )

// func gl212main() {
// 	// func main() {
// 	runtime.LockOSThread()
// 	window := initGlfw212()
// 	defer glfw.Terminate()

// 	initOpenGL212()

// 	gmap := initMap()

// 	// vao := makeVao(triangle)
// 	for !window.ShouldClose() {
// 		// fmt.Println("draw for")
// 		tr := base.NewTimeRecorder()
// 		drawgl212(window, gmap)
// 		tr.Output("draw")
// 	}
// }

// func initMap() *mapping.Map {
// 	gmap := mapping.NewMap(draw.GL)
// 	gmap.Open("c:/temp/JBNTBHTB.gmp") // railway chinapnt_84 County_R Provinces JBNTBHTB  dltb
// 	gmap.Prepare(width, height)
// 	// gmap.Zoom(2.1)
// 	return gmap
// }

// // var InnerContour [4][3]float64 = [4][3]float64{[3]float64{-1, 1, 0},
// // 	[3]float64{1, 1, 0},
// // 	[3]float64{1, -1, 0},
// // 	[3]float64{-1, -1, 0}}

// var polygon [6][3]float64 = [6][3]float64{
// 	{-0.05, 0, 0},
// 	{-0.1, 0.1, 0},
// 	{0.1, 0.1, 0},
// 	{0.05, 0, 0},

// 	{0.1, -0.1, 0},
// 	{-0.1, -0.1, 0}}

// // {-0.1, 0.1, 0}}

// // var polygon [7][3]float64 = [7][3]float64{[3]float64{-0.1, 0.1, 0},
// // 	[3]float64{0.1, 0.1, 0},
// // 	[3]float64{0.05, 0, 0},
// // 	[3]float64{0.1, -0.1, 0},
// // 	[3]float64{-0.1, -0.1, 0},
// // 	[3]float64{0.05, 0, 0},
// // 	[3]float64{-0.1, 0.1, 0}}

// // type PolygonData struct {
// // 	BeginCount    int
// // 	VertexCount   int
// // 	EndCount      int
// // 	ErrorCount    int
// // 	EdgeFlagCount int
// // 	CombineCount  int

// // 	Vertices []VertexData
// // }

// // type VertexData struct {
// // 	Location    [3]float64
// // 	VertexHits  int
// // 	CombineHits int
// // }

// // // Test shape is a square with a square hole inside.
// // var OuterContour [4][3]float64 = [4][3]float64{[3]float64{-2, 2, 0},
// // 	[3]float64{-2, -2, 0},
// // 	[3]float64{2, -2, 0},
// // 	[3]float64{2, 2, 0}}

// // var InnerContour [4][3]float64 = [4][3]float64{[3]float64{-1, 1, 0},
// // 	[3]float64{1, 1, 0},
// // 	[3]float64{1, -1, 0},
// // 	[3]float64{-1, -1, 0}}

// // // Pentagram with crossing edges. Invokes the combine callback.
// // var StarContour [5][3]float64 = [5][3]float64{[3]float64{0, 1, 0},
// // 	[3]float64{-1, -1, 0},
// // 	[3]float64{1, 0, 0},
// // 	[3]float64{-1, 0, 0},
// // 	[3]float64{1, -1, 0}}

// func tessBeginNilHandler(tessType uint32, polygonData interface{}) {
// 	gl.Begin(tessType)
// 	// fmt.Println("begin:", tessType)
// }

// func tessVertexNilHandler(vertexData interface{}, polygonData interface{}) {
// 	// fmt.Println("vertex:", vertexData)
// 	// glVertex3dv(vertex)
// 	data := vertexData.([3]float64)
// 	gl.Vertex3d(data[0], data[1], data[2])
// }

// func tessCombineDataHandler(coords [3]float64, vertexData [4]interface{},
// 	weight [4]float32, polygonData interface{}) (outData interface{}) {
// 	fmt.Println("tessCombineDataHandler:", coords, weight, polygonData)
// 	gl.Vertex3d(coords[0], coords[1], coords[2])
// 	return
// }

// func tessEndNilHandler(polygonData interface{}) {
// 	gl.End()
// 	// fmt.Println("end")
// }

// func tessErrorNilHandler(errno uint32, polygonData interface{}) {
// 	fmt.Println("error:", errno)
// }

// func drawPolygon2() {
// 	tess := glu.NewTess()
// 	tess.SetBeginCallback(tessBeginNilHandler)
// 	tess.SetVertexCallback(tessVertexNilHandler)
// 	tess.SetEndCallback(tessEndNilHandler)
// 	tess.SetErrorCallback(tessErrorNilHandler)
// 	tess.SetCombineCallback(tessCombineDataHandler)

// 	// tess.Normal(0, 0, 1)

// 	tess.BeginPolygon(&polygon)

// 	tess.BeginContour()
// 	for i, _ := range polygon {
// 		tess.Vertex2(polygon[i], unsafe.Pointer(&polygon[i][0]))
// 	}
// 	tess.EndContour()

// 	tess.EndPolygon()
// 	tess.Delete()
// }

// var polygonf [6][3]float32 = [6][3]float32{
// 	{-0.05, 0, 0},
// 	{-0.1, 0.1, 0},
// 	{0.1, 0.1, 0},
// 	{0.05, 0, 0},

// 	{0.1, -0.1, 0},
// 	{-0.1, -0.1, 0}}

// func drawPolygon() {
// 	gl.Begin(gl.POLYGON)
// 	for _, v := range polygonf {
// 		gl.Vertex3f(v[0], v[1], v[2])
// 	}
// 	gl.End()
// }

// var triangle1 [3][3]float32 = [3][3]float32{
// 	{-0.05, 0, 0},
// 	{-0.1, 0.1, 0},
// 	{0.05, 0, 0}}
// var triangle2 [3][3]float32 = [3][3]float32{
// 	{-0.1, 0.1, 0},
// 	{0.05, 0, 0},
// 	{0.1, 0.1, 0}}
// var triangle3 [3][3]float32 = [3][3]float32{
// 	{-0.05, 0, 0},
// 	{-0.1, -0.1, 0},
// 	{0.05, 0, 0}}
// var triangle4 [3][3]float32 = [3][3]float32{
// 	{-0.05, 0, 0},
// 	{0.1, -0.1, 0},
// 	{0.05, 0, 0}}

// func drawtri(tri [3][3]float32) {
// 	gl.Begin(gl.TRIANGLES)
// 	for _, v := range tri {
// 		gl.Vertex3f(v[0], v[1], v[2])
// 	}
// 	gl.End()
// }

// func drawPolygon3() {
// 	drawtri(triangle1)
// 	drawtri(triangle2)
// 	drawtri(triangle3)
// 	drawtri(triangle4)
// }

// var tris [3][3]float32 = [3][3]float32{
// 	{0, 0.1, 0},
// 	{0.1, -0.1, 0},
// 	{-0.1, -0.1, 0}}

// func drawtri3() {
// 	gl.Begin(gl.TRIANGLES)
// 	for _, v := range tris {
// 		gl.Vertex3f(v[0], v[1], v[2])
// 	}
// 	gl.End()
// }

// func drawgl212(window *glfw.Window, gmap *mapping.Map) {
// 	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
// 	gl.Clear(gl.COLOR_BUFFER_BIT)

// 	gl.LineWidth(2)
// 	gl.Color3f(1, 0.5, 0)

// 	drawMap(gmap)
// 	// drawLine()
// 	// for i := 0; i < 100000; i++ {
// 	// 	drawtri3()
// 	// }

// 	glfw.PollEvents()
// 	window.SwapBuffers()
// }

// func drawLine() {
// 	// gl.Begin(gl.LINE_STRIP)
// 	// gl.Vertex3f(0.005, -0.005, 0)
// 	// gl.Vertex3f(0.005, 0.005, 0)
// 	// gl.Vertex3f(-0.005, 0.005, 0)
// 	// gl.Vertex3f(-0.005, -0.005, 0)
// 	// gl.Vertex3f(0.005, -0.005, 0)
// 	// gl.End()

// 	for i := 100; i < 2000; i++ {
// 		gl.Begin(gl.LINE_STRIP)
// 		d := float32(0.00001)
// 		c := float32(i) * d
// 		gl.Vertex3f(c+d, c-d, 0)
// 		gl.Vertex3f(c+d, c+d, 0)
// 		gl.Vertex3f(c-d, c+d, 0)
// 		gl.Vertex3f(c-d, c-d, 0)
// 		gl.Vertex3f(c+d, c-d, 0)
// 		gl.End()
// 	}

// }

// func drawMap(gmap *mapping.Map) {
// 	gmap.Draw()
// }

// // initOpenGL 初始化 OpenGL 并且返回一个初始化了的程序。
// func initOpenGL212() {
// 	if err := gl.Init(); err != nil {
// 		panic(err)
// 	}
// 	version := gl.GoStr(gl.GetString(gl.VERSION))
// 	fmt.Println("OpenGL Version:", version)
// 	return
// }

// // initGlfw 初始化 glfw 并且返回一个可用的窗口。
// func initGlfw212() *glfw.Window {
// 	if err := glfw.Init(); err != nil {
// 		panic(err)
// 	}
// 	glfw.WindowHint(glfw.Resizable, glfw.True)
// 	glfw.WindowHint(glfw.ContextVersionMajor, 2) // OR 2
// 	glfw.WindowHint(glfw.ContextVersionMinor, 1)
// 	window, err := glfw.CreateWindow(width, height, "test", nil, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	window.MakeContextCurrent()
// 	return window
// }

// // ===========================================

// // void CALLBACK vertexCallback(GLvoid* vertex)
// // {
// // 	GLdouble* pt;

// // 	int numb;
// // 	pt = (GLdouble*)vertex;

// // 	glColor3f(0.8, 0.8, 0.8);

// // 	glVertex3d(pt[0],pt[1],pt[2]);
// // }

// // void CALLBACK beginCallback(GLenum type)
// // {
// // 	glBegin(type);
// // }

// // void CALLBACK endCallback()
// // {
// // 	glEnd();
// // }

// // void CALLBACK errorCallback(GLenum errorCode)
// // {
// // 	const GLubyte * estring;
// // 	//打印错误类型
// // 	estring = gluErrorString(errorCode);
// // 	fprintf(stderr, "Tessellation Error: %s/n", estring);
// // 	exit(0);
// // }
