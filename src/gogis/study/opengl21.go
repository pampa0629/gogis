package main

// import (
// 	"fmt"
// 	"gogis/base"
// 	"runtime"

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

// func gl21main() {
// 	// func main() {
// 	runtime.LockOSThread()
// 	window := initGlfw21()
// 	defer glfw.Terminate()

// 	initOpenGL21()
// 	// vao := makeVao(triangle)
// 	for !window.ShouldClose() {
// 		// fmt.Println("draw for")
// 		tr := base.NewTimeRecorder()
// 		drawgl21(window)
// 		tr.Output("draw")
// 	}
// }

// // var (
// // 	polygon = [][3]float64{
// // 		[3]float64{-0.1, 0.1, 0},
// // 		[3]float64{0.1, 0.1, 0},
// // 		[3]float64{0.05, 0, 0},
// // 		[3]float64{0.1, -0.1, 0},
// // 		[3]float64{-0.1, -0.1, 0},
// // 		[3]float64{0.05, 0, 0},
// // 		[3]float64{-0.1, 0.1, 0}}
// // )

// type PolygonData struct {
// 	BeginCount    int
// 	VertexCount   int
// 	EndCount      int
// 	ErrorCount    int
// 	EdgeFlagCount int
// 	CombineCount  int

// 	Vertices []VertexData
// }

// type VertexData struct {
// 	Location    [3]float64
// 	VertexHits  int
// 	CombineHits int
// }

// // Test shape is a square with a square hole inside.
// var OuterContour [4][3]float64 = [4][3]float64{[3]float64{-2, 2, 0},
// 	[3]float64{-2, -2, 0},
// 	[3]float64{2, -2, 0},
// 	[3]float64{2, 2, 0}}

// var InnerContour [4][3]float64 = [4][3]float64{[3]float64{-1, 1, 0},
// 	[3]float64{1, 1, 0},
// 	[3]float64{1, -1, 0},
// 	[3]float64{-1, -1, 0}}

// // Pentagram with crossing edges. Invokes the combine callback.
// var StarContour [5][3]float64 = [5][3]float64{[3]float64{0, 1, 0},
// 	[3]float64{-1, -1, 0},
// 	[3]float64{1, 0, 0},
// 	[3]float64{-1, 0, 0},
// 	[3]float64{1, -1, 0}}

// // func tessBeginNilHandler(tessType uint32, polygonData interface{}) {
// // 	gl.Begin(tessType)
// // }

// // func tessVertexNilHandler(vertexData interface{}, polygonData interface{}) {
// // 	fmt.Println("vertex:", vertexData)
// // }

// // func tessEndNilHandler(polygonData interface{}) {
// // 	gl.End()
// // }

// // func tessErrorNilHandler(errno uint32, polygonData interface{}) {
// // 	fmt.Println("error:", errno)
// // }

// func drawPolygon21() {
// 	poly := new(PolygonData)

// 	for _, v := range OuterContour {
// 		poly.Vertices = append(poly.Vertices, VertexData{Location: v})
// 	}
// 	for _, v := range InnerContour {
// 		poly.Vertices = append(poly.Vertices, VertexData{Location: v})
// 	}

// 	tess := glu.NewTess()
// 	// tess.SetBeginCallback(tessBeginNilHandler)
// 	// tess.SetVertexCallback(tessVertexNilHandler)
// 	// tess.SetEndCallback(tessEndNilHandler)
// 	// tess.SetErrorCallback(tessErrorNilHandler)
// 	tess.SetBeginCallback(tessBeginDataHandler)
// 	tess.SetVertexCallback(tessVertexDataHandler)
// 	tess.SetEndCallback(tessEndDataHandler)
// 	tess.SetErrorCallback(tessErrorDataHandler)
// 	tess.SetEdgeFlagCallback(tessEdgeFlagDataHandler)
// 	tess.SetCombineCallback(tessCombineDataHandler)

// 	tess.Normal(0, 0, 1)

// 	tess.BeginPolygon(nil)
// 	tess.BeginContour()

// 	for v := 0; v < 4; v += 1 {
// 		tess.Vertex(poly.Vertices[v].Location, &poly.Vertices[v])
// 	}

// 	tess.EndContour()
// 	tess.BeginContour()

// 	for v := 4; v < 8; v += 1 {
// 		tess.Vertex(poly.Vertices[v].Location, &poly.Vertices[v])
// 	}

// 	tess.EndContour()
// 	tess.EndPolygon()

// 	expectedTriangles := 8
// 	// There are a total of 24 edges, 8 of which are not edges. This means
// 	// the EdgeFlag must be toggled to true 8 times.
// 	expectedEdges := 8

// 	checkPoly(poly, 1, expectedTriangles*3, 1, 0, expectedEdges, 0)

// 	tess.Delete()
// }

// // func drawPolygon() {
// // 	gl.Begin(gl.POLYGON)

// // 	gl.Vertex3f(-0.1, 0.1, 0.0)
// // 	gl.Vertex3f(0.1, 0.1, 0.0)
// // 	gl.Vertex3f(0.05, 0, 0.0)
// // 	gl.Vertex3f(0.1, -0.1, 0.0)
// // 	gl.Vertex3f(-0.1, -0.1, 0.0)
// // 	gl.Vertex3f(-0.05, 0, 0.0)
// // 	gl.Vertex3f(-0.1, 0.1, 0.0)

// // 	gl.End()
// // }

// func drawgl21(window *glfw.Window) {
// 	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
// 	gl.Clear(gl.COLOR_BUFFER_BIT)

// 	gl.LineWidth(2)
// 	gl.Color3f(1, 0.5, 0)

// 	for i := 0; i < 10; i++ {
// 		drawPolygon()
// 	}

// 	glfw.PollEvents()
// 	window.SwapBuffers()
// }

// // initOpenGL 初始化 OpenGL 并且返回一个初始化了的程序。
// func initOpenGL21() {
// 	if err := gl.Init(); err != nil {
// 		panic(err)
// 	}
// 	version := gl.GoStr(gl.GetString(gl.VERSION))
// 	fmt.Println("OpenGL Version:", version)
// 	return
// }

// // initGlfw 初始化 glfw 并且返回一个可用的窗口。
// func initGlfw21() *glfw.Window {
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

// func tessBeginDataHandler(tessType uint32, polygonData interface{}) {
// 	polygonDataPtr := polygonData.(*PolygonData)
// 	polygonDataPtr.BeginCount += 1
// }

// func tessVertexDataHandler(vertexData interface{}, polygonData interface{}) {
// 	polygonDataPtr := polygonData.(*PolygonData)
// 	polygonDataPtr.VertexCount += 1

// 	vertexDataPtr := vertexData.(*VertexData)
// 	vertexDataPtr.VertexHits += 1
// }

// func tessEndDataHandler(polygonData interface{}) {
// 	polygonDataPtr := polygonData.(*PolygonData)
// 	polygonDataPtr.EndCount += 1
// }

// func tessErrorDataHandler(errno uint32, polygonData interface{}) {
// 	polygonDataPtr := polygonData.(*PolygonData)
// 	polygonDataPtr.ErrorCount += 1
// }

// func tessEdgeFlagDataHandler(flag bool, polygonData interface{}) {
// 	polygonDataPtr := polygonData.(*PolygonData)

// 	if flag {
// 		polygonDataPtr.EdgeFlagCount += 1
// 	}
// }

// func tessCombineDataHandler21(coords [3]float64,
// 	vertexData [4]interface{},
// 	weight [4]float32,
// 	polygonData interface{}) (outData interface{}) {

// 	polygonDataPtr := polygonData.(*PolygonData)
// 	polygonDataPtr.CombineCount += 1

// 	for _, v := range vertexData {
// 		vertexDataPtr := v.(*VertexData)
// 		vertexDataPtr.CombineHits += 1
// 	}

// 	newVertex := VertexData{Location: coords}
// 	polygonDataPtr.Vertices = append(polygonDataPtr.Vertices, newVertex)

// 	return &polygonDataPtr.Vertices[len(polygonDataPtr.Vertices)-1]
// }

// func checkPoly(poly *PolygonData, expectedBegins, expectedVertices,
// 	expectedEnds, expectedErrors, expectedEdges, expectedCombines int) {
// 	if poly.BeginCount != expectedBegins {
// 		fmt.Println("Expected BeginCount == %v, got %v\n",
// 			expectedBegins,
// 			poly.BeginCount)
// 	}
// 	if poly.VertexCount != expectedVertices {
// 		fmt.Println("Expected VertexCount == %v, got %v\n",
// 			expectedVertices,
// 			poly.VertexCount)
// 	}
// 	if poly.EndCount != expectedEnds {
// 		fmt.Println("Expected EndCount == %v, got %v\n",
// 			expectedEnds,
// 			poly.EndCount)
// 	}
// 	if poly.ErrorCount != expectedErrors {
// 		fmt.Println("Expected ErrorCount == %v, got %v\n",
// 			expectedErrors,
// 			poly.ErrorCount)
// 	}
// 	if poly.EdgeFlagCount != expectedEdges {
// 		fmt.Println("Expected EdgeFlagCount == %v, got %v\n",
// 			expectedEdges,
// 			poly.EdgeFlagCount)
// 	}
// 	if poly.CombineCount != expectedCombines {
// 		fmt.Println("Expected CombineCount == %v, got %v\n",
// 			expectedCombines,
// 			poly.CombineCount)
// 	}
// }
