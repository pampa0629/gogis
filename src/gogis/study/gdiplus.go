package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	. "github.com/tryor/gdiplus"

	. "github.com/tryor/winapi"

	. "github.com/tryor/winapi/gdi"
)

func abortf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format, a...)
	os.Exit(1)
}

func abortErrNo(funcname string, err error) {
	abortf("%s failed: %d %s\n", funcname, err, err)
}

var (
	mh HINSTANCE
)

//var appn *Application

const Width, Height = 600, 450

var (
	hostDC      HDC
	bufferDC    HDC
	hbitmap     HANDLE
	bitmapLayer *Bitmap
	graphics    *Graphics
	gpToken     ULONG_PTR

	hostGraphics *Graphics
)

func startupGdiplus() {
	status, err := Startup(&gpToken, nil, nil)
	fmt.Println("Startup.status:", status)
	fmt.Println("Startup.err:", err)
}

var paths []*GraphicsPath

func drawShape(mx, my int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Runtime error caught: %v\n", r)
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				log.Print(file, line)
			}
		}
	}()

	for i := 0; i < 1; i++ {
		drawShapeA(mx+i, my+i)
	}

	hostGraphics.DrawImageI3(bitmapLayer, 0, 0, INT(bitmapLayer.GetWidth()), INT(bitmapLayer.GetHeight()))

	BitBlt(hostDC, 0, 0, Width, Height, bufferDC, 0, 0, SRCCOPY)
}

func drawShapeA(mx, my int) {
	pen, err := NewPen(Color{Argb: Red}, 3)
	defer pen.Release()
	if err != nil {
		fmt.Println("NewPen.err:", err)
		return
	}

	brush, err := NewSolidBrush(NewColor3(255, 200, 200, 100))
	defer brush.Release()
	if err != nil {
		fmt.Println("NewSolidBrush.err:", err)
		return
	}

	fontbrush, err := NewSolidBrush(NewColor3(255, 200, 0, 100))
	defer fontbrush.Release()
	if err != nil {
		fmt.Println("NewSolidBrush.err:", err)
		return
	}

	path, err := NewGraphicsPath()
	if err != nil {
		fmt.Println("NewGraphicsPath.err:", err)
		return
	}

	rect := &RectF{REAL(mx), REAL(my), 100, 30}
	path.AddRectangle(rect)

	paths = append(paths, path)

	graphics.FillPath(brush, path)
	if graphics.LastResult != Ok {
		fmt.Println("graphics.FillPath.err:", err, graphics.LastResult)
	}
	graphics.DrawPath(pen, path)
	if graphics.LastResult != Ok {
		fmt.Println("graphics.DrawPath.err:", err, graphics.LastResult)
	}

	path.IsVisible(REAL(mx), REAL(my), graphics)
	if path.LastResult != Ok {
		fmt.Println("path.IsVisible.err:", path.LastResult)
	}

	path.IsOutlineVisible(REAL(mx), REAL(my), pen, graphics)
	if path.LastResult != Ok {
		fmt.Println("path.IsOutlineVisible.err:", path.LastResult)
	}

	path.Reset()
	if path.LastResult != Ok {
		fmt.Println("path.Reset.err:", path.LastResult)
	}

	//family, _ := NewFontFamily("Arial", nil)
	//defer family.Release()
	//font, _ := NewFont(family, 15, FontStyleRegular, UnitPixel)
	//defer font.Release()

	//familyName string, emSize REAL, style FontStyle, unit Unit, fontCollection IFontCollection
	font, _ := NewFont2("Arial", 15, FontStyleRegular, UnitPixel, nil)
	defer font.Release()
	fmt.Println("NewFont2.err:", font.LastResult)
	fmt.Println("NewFont2.font:", font)

	text := fmt.Sprintf("test:%v,%v", 2, 3)
	graphics.DrawString(text, font, rect, nil, fontbrush)

}

////appn = NewApplication(hwnd, backBuffer)
func graphics_example(hwnd HWND) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Runtime error caught: %v\n", r)
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				log.Print(file, line)
			}
		}
	}()

	startupGdiplus()

	paths = make([]*GraphicsPath, 0)

	hostDC = GetDC(hwnd)

	// 创建双缓冲
	bufferDC = CreateCompatibleDC(HWND(hostDC))
	hbitmap = CreateCompatibleBitmap(hostDC, Width, Height)
	SelectObject(bufferDC, hbitmap)
	DeleteObject(hbitmap)

	var err error
	hostGraphics, err = FromHDC(bufferDC)
	//	defer graphics.Release()
	fmt.Println("FromHDC.hostGraphics:", hostGraphics)
	fmt.Println("FromHDC.status:", hostGraphics.LastResult)
	fmt.Println("FromHDC.err:", err)

	hostGraphics.SetPageUnit(UnitPixel)
	hostGraphics.SetSmoothingMode(SmoothingModeHighQuality)
	hostGraphics.SetTextRenderingHint(TextRenderingHintClearTypeGridFit)
	//hostGraphics.SetTextRenderingHint(TextRenderingHintAntiAlias)
	hostGraphics.Clear(Color{Olive})

	bitmapLayer, _ = NewBitmap3(1000, 1000, PixelFormat32bppARGB)
	if bitmapLayer.LastError != nil {
		fmt.Println("NewBitmap3.err:", bitmapLayer.LastError)
		return
	}
	graphics, _ = FromImage(bitmapLayer)
	if graphics.LastError != nil {
		fmt.Println("FromImage.err:", graphics.LastError)
		return
	}
	graphics.SetPageUnit(UnitPixel)
	graphics.SetSmoothingMode(SmoothingModeHighQuality)
	graphics.SetTextRenderingHint(TextRenderingHintClearTypeGridFit)
	graphics.Clear(Color{Olive})

	pen, err := NewPen(Color{Argb: Red}, 3)
	defer pen.Release()
	fmt.Println("NewPen.err:", err)
	if err != nil {
		return
	}

	brush, err := NewSolidBrush(NewColor3(255, 200, 200, 100))
	defer brush.Release()
	fmt.Println("NewSolidBrush.brush:", brush)
	if err != nil {
		return
	}
	//	return
	gpath, _ := NewGraphicsPath()
	defer gpath.Release()
	fmt.Println("gpath.status:", gpath.LastResult)
	fmt.Println("gpath.err:", gpath.LastError)

	gpath.AddRectangle(NewRectF(10, 10, 250, 100))
	fmt.Println("gpath.AddRectangle.status:", gpath.LastResult)
	fmt.Println("gpath.AddRectangle.err:", gpath.LastError)

	family, _ := NewFontFamily("宋体", nil) //"Courier New"
	defer family.Release()
	gpath.AddString("测试", family, FontStyleBold|FontStyleItalic, 80, &RectF{25, 25, 0.0, 0.0}, nil)
	fmt.Println("gpath.AddString.status:", gpath.LastResult)
	fmt.Println("gpath.AddString.err:", gpath.LastError)

	graphics.FillPath(brush, gpath)
	fmt.Println("FillPath.status:", graphics.LastResult)
	fmt.Println("FillPath.err:", graphics.LastError)

	graphics.DrawPath(pen, gpath)
	fmt.Println("DrawPath.status:", graphics.LastResult)
	fmt.Println("DrawPath.err:", graphics.LastError)

	font, _ := NewFont(family, 50, FontStyleBold|FontStyleItalic, UnitPixel)
	fmt.Println("NewFont.status:", font.LastResult)
	fmt.Println("NewFont.err:", font.LastError)
	defer font.Release()

	rect := &RectF{10, 150, 0, 0}
	rect, codepointsFitted, linesFilled, _ := graphics.MeasureString("测试", font, rect, nil)
	fmt.Println("MeasureString.status:", graphics.LastResult)
	fmt.Println("MeasureString.err:", graphics.LastError)
	fmt.Println("MeasureString.linesFilled:", linesFilled)
	fmt.Println("MeasureString.codepointsFitted:", codepointsFitted)
	fmt.Println("MeasureString.rect:", rect)

	graphics.DrawRectangle2(pen, rect)
	graphics.DrawString("测试", font, rect, nil, brush)
	fmt.Println("DrawString.status:", graphics.LastResult)
	fmt.Println("DrawString.err:", graphics.LastError)

	appPath, _ := os.Getwd()
	bitmap, _ := NewBitmap(appPath + "/penguins.jpg")
	//	bitmap, _ := NewBitmap(appPath + "/i2_select.png")
	defer bitmap.Release()
	fmt.Println("NewBitmap.status:", bitmap.LastResult)
	fmt.Println("NewBitmap.err:", bitmap.LastError)
	fmt.Println("NewBitmap.bitmap:", bitmap)

	//	graphics.DrawImageI(bitmap, 200, 150)
	//		graphics.DrawImageI3(bitmap, 200, 150, INT(bitmap.GetWidth()), INT(bitmap.GetHeight()))
	//	graphics.DrawImageI3(bitmap, 200, 150, 200, 200)
	//graphics.DrawImageI6(bitmap, 200, 150, 0, 0, INT(bitmap.GetWidth()), INT(bitmap.GetHeight()), UnitPixel)
	graphics.DrawImageI6(bitmap, 200, 150, 100, 100, 200, 200, UnitPixel)
	fmt.Println("DrawImageI4.status:", graphics.LastResult)
	fmt.Println("DrawImageI4.err:", graphics.LastError)

	graphics.DrawDriverString("测试TEST223", font, brush, &PointF{150, 150}, DriverStringOptionsRealizedAdvance|DriverStringOptionsVertical|DriverStringOptionsCmapLookup|DriverStringOptionsCompensateResolution, nil)
	fmt.Println("DrawDriverString.status:", graphics.LastResult)
	fmt.Println("DrawDriverString.err:", graphics.LastError)

	layerCanvas, _ := NewBitmap3(100, 200, PixelFormat32bppARGB)
	fmt.Println("NewBitmap3.status:", layerCanvas.LastResult)
	fmt.Println("NewBitmap3.err:", layerCanvas.LastError)
	fmt.Println("NewBitmap3.bitmap:", layerCanvas)

	points := []PointF{PointF{100, 100}, PointF{100, 200}, PointF{150, 280}, PointF{200, 300}, PointF{250, 200}, PointF{200, 130}, PointF{100, 100}}
	graphics.DrawCurve(pen, points)
	fmt.Println("DrawCurve.status:", graphics.LastResult)
	fmt.Println("DrawCurve.err:", graphics.LastError)

	//	graphics.DrawLines(pen, points)
	//	fmt.Println("DrawLines.status:", graphics.LastResult)
	//	fmt.Println("DrawLines.err:", graphics.LastError)

	BitBlt(hostDC, 0, 0, Width, Height, bufferDC, 0, 0, SRCCOPY)
}

// WinProc called by windows to notify us of all windows events we might be interested in.
func WndProc(hwnd HWND, msg UINT, wparam WPARAM, lparam LPARAM) (rc uintptr) {
	switch msg {
	case WM_CREATE:
		rc = DefWindowProcW(hwnd, msg, wparam, lparam)
	case WM_CLOSE:
		Shutdown(gpToken)
		DestroyWindow(hwnd)
	case WM_COMMAND:
		switch HANDLE(lparam) {
		default:
			rc = DefWindowProcW(hwnd, msg, wparam, lparam)
		}
	case WM_PAINT:
		BitBlt(hostDC, 0, 0, Width, Height, bufferDC, 0, 0, SRCCOPY)
		rc = DefWindowProcW(hwnd, msg, wparam, lparam)
	case WM_DESTROY:
		PostQuitMessage(0)
	case WM_MOUSEMOVE:
		//appn.TrackMouseMoveEvent(int(LOWORD(lparam)), int(HIWORD(lparam)), easydraw.MButton(wparam))
	case WM_LBUTTONDOWN, WM_RBUTTONDOWN:
		drawShape(int(LOWORD(INT(lparam))), int(HIWORD(INT(lparam))))
		//appn.TrackMousePressEvent(int(LOWORD(lparam)), int(HIWORD(lparam)), easydraw.MButton(wparam))
	case WM_LBUTTONUP, WM_RBUTTONUP:
		//appn.TrackMouseReleaseEvent(int(LOWORD(lparam)), int(HIWORD(lparam)), easydraw.MButton(wparam))
	case WM_LBUTTONDBLCLK:
		//onMouseLeftDoubleClick
	case WM_SETCURSOR:
		//if appn.SetCursor() {
		//	return
		//} else {
		//	return DefWindowProcW(hwnd, msg, wparam, lparam)
		//}
		return DefWindowProcW(hwnd, msg, wparam, lparam)
	case WM_KEYDOWN:
		//appn.TrackKeyPressEvent(int(wparam))
	case WM_KEYUP:
		//onKeyRelease(wParam)

	case WM_CHAR:
		//		s := string(int(wparam))
		//		enc := mahonia.NewDecoder("utf-8")
		//		v := enc.ConvertString(s)
		//		var c rune
		//		if v != "" {
		//			runes := bytes.Runes([]byte(v))
		//			c = runes[0]
		//		}
		//		fmt.Println("c:", c, string(c))
		//appn.TrackKeyCharEvent(c)

	default:
		rc = DefWindowProcW(hwnd, msg, wparam, lparam)
	}
	return
}

func rungui() int {
	var e error

	// GetModuleHandle
	mh, e = GetModuleHandle("")
	if e != nil {
		abortErrNo("GetModuleHandle", e)
	}

	// Get icon we're going to use.
	myicon, e := LoadIcon(0, IDI_APPLICATION)
	if e != nil {
		abortErrNo("LoadIcon", e)
	}

	// Get cursor we're going to use.
	mycursor, e := LoadCursor(0, IDC_ARROW)
	if e != nil {
		abortErrNo("LoadCursor", e)
	}

	// Create callback
	wproc := syscall.NewCallback(WndProc)

	// RegisterClassEx
	wcname := "my Window Class" //syscall.StringToUTF16Ptr(wcname)
	var wc Wndclassex
	wc.Size = uint32(unsafe.Sizeof(wc))
	wc.WndProc = wproc
	wc.Instance = HINSTANCE(mh)
	wc.Icon = myicon
	wc.Cursor = mycursor
	wc.Background = COLOR_BTNFACE + 1
	wc.MenuName = nil
	wc.ClassName = syscall.StringToUTF16Ptr(wcname)
	wc.IconSm = myicon
	if _, e := RegisterClassExW(&wc); e != nil {
		abortErrNo("RegisterClassEx", e)
	}

	// CreateWindowEx
	wh, e := CreateWindowExW(
		WS_EX_CLIENTEDGE,
		wcname,
		"My window",
		WS_OVERLAPPEDWINDOW|WS_VISIBLE,
		CW_USEDEFAULT, CW_USEDEFAULT,
		Width+20, Height+40+25,
		0, 0, HINSTANCE(mh), 0)

	fmt.Printf("e %v\n", e)
	if e != nil {
		abortErrNo("CreateWindowEx", e)
	}
	fmt.Printf("main window handle is %x\n", wh)

	// ShowWindow
	ShowWindow(wh, SW_SHOWDEFAULT)

	if e := UpdateWindow(wh); e != nil {
		abortErrNo("UpdateWindow", e)
	}

	graphics_example(wh)

	// Process all windows messages until WM_QUIT.
	var m Msg
	for {
		r, e := GetMessage(&m, 0, 0, 0)
		if e != nil {
			abortErrNo("GetMessage", e)
		}
		if r == 0 {
			break
		}
		TranslateMessage(&m)
		DispatchMessageW(&m)
	}
	return int(m.Wparam)
}

func wholemain() {
	//	FreeConsole()
	rc := rungui()
	os.Exit(rc)
}

func gdimain() {
	startupGdiplus()
	fmt.Println("============startupGdiplus() end============")

	bitmapLayer, _ = NewBitmap3(1000, 1000, PixelFormat32bppARGB)
	if bitmapLayer.LastError != nil {
		fmt.Println("NewBitmap3.err:", bitmapLayer.LastError)
		return
	}
	graphics, _ := FromImage(bitmapLayer)
	if graphics.LastError != nil {
		fmt.Println("FromImage.err:", graphics.LastError)
		return
	}
	graphics.SetPageUnit(UnitPixel)
	graphics.SetSmoothingMode(SmoothingModeHighQuality)
	graphics.SetTextRenderingHint(TextRenderingHintClearTypeGridFit)
	graphics.Clear(Color{Olive})

	pen, err := NewPen(Color{Argb: Red}, 3)
	defer pen.Release()
	fmt.Println("NewPen.err:", err)
	if err != nil {
		return
	}

	brush, err := NewSolidBrush(NewColor3(255, 200, 200, 100))
	defer brush.Release()
	fmt.Println("NewSolidBrush.brush:", brush)
	if err != nil {
		return
	}
	//	return
	gpath, _ := NewGraphicsPath()
	defer gpath.Release()
	fmt.Println("gpath.status:", gpath.LastResult)
	fmt.Println("gpath.err:", gpath.LastError)

	gpath.AddRectangle(NewRectF(10, 10, 250, 100))
	fmt.Println("gpath.AddRectangle.status:", gpath.LastResult)
	fmt.Println("gpath.AddRectangle.err:", gpath.LastError)

	family, _ := NewFontFamily("宋体", nil) //"Courier New"
	defer family.Release()
	gpath.AddString("测试", family, FontStyleBold|FontStyleItalic, 80, &RectF{25, 25, 0.0, 0.0}, nil)
	fmt.Println("gpath.AddString.status:", gpath.LastResult)
	fmt.Println("gpath.AddString.err:", gpath.LastError)

	graphics.FillPath(brush, gpath)
	fmt.Println("FillPath.status:", graphics.LastResult)
	fmt.Println("FillPath.err:", graphics.LastError)

	graphics.DrawPath(pen, gpath)
	fmt.Println("DrawPath.status:", graphics.LastResult)
	fmt.Println("DrawPath.err:", graphics.LastError)

	font, _ := NewFont(family, 50, FontStyleBold|FontStyleItalic, UnitPixel)
	fmt.Println("NewFont.status:", font.LastResult)
	fmt.Println("NewFont.err:", font.LastError)
	defer font.Release()

	rect := &RectF{10, 150, 0, 0}
	rect, codepointsFitted, linesFilled, _ := graphics.MeasureString("测试", font, rect, nil)
	fmt.Println("MeasureString.status:", graphics.LastResult)
	fmt.Println("MeasureString.err:", graphics.LastError)
	fmt.Println("MeasureString.linesFilled:", linesFilled)
	fmt.Println("MeasureString.codepointsFitted:", codepointsFitted)
	fmt.Println("MeasureString.rect:", rect)

	graphics.DrawRectangle2(pen, rect)
	graphics.DrawString("测试", font, rect, nil, brush)
	fmt.Println("DrawString.status:", graphics.LastResult)
	fmt.Println("DrawString.err:", graphics.LastError)

	appPath, _ := os.Getwd()
	bitmap, _ := NewBitmap(appPath + "/penguins.jpg")
	//	bitmap, _ := NewBitmap(appPath + "/i2_select.png")
	defer bitmap.Release()
	fmt.Println("NewBitmap.status:", bitmap.LastResult)
	fmt.Println("NewBitmap.err:", bitmap.LastError)
	fmt.Println("NewBitmap.bitmap:", bitmap)

	//	graphics.DrawImageI(bitmap, 200, 150)
	//		graphics.DrawImageI3(bitmap, 200, 150, INT(bitmap.GetWidth()), INT(bitmap.GetHeight()))
	//	graphics.DrawImageI3(bitmap, 200, 150, 200, 200)
	//graphics.DrawImageI6(bitmap, 200, 150, 0, 0, INT(bitmap.GetWidth()), INT(bitmap.GetHeight()), UnitPixel)
	graphics.DrawImageI6(bitmap, 200, 150, 100, 100, 200, 200, UnitPixel)
	fmt.Println("DrawImageI4.status:", graphics.LastResult)
	fmt.Println("DrawImageI4.err:", graphics.LastError)

	graphics.DrawDriverString("测试TEST223", font, brush, &PointF{150, 150}, DriverStringOptionsRealizedAdvance|DriverStringOptionsVertical|DriverStringOptionsCmapLookup|DriverStringOptionsCompensateResolution, nil)
	fmt.Println("DrawDriverString.status:", graphics.LastResult)
	fmt.Println("DrawDriverString.err:", graphics.LastError)

	layerCanvas, _ := NewBitmap3(100, 200, PixelFormat32bppARGB)
	fmt.Println("NewBitmap3.status:", layerCanvas.LastResult)
	fmt.Println("NewBitmap3.err:", layerCanvas.LastError)
	fmt.Println("NewBitmap3.bitmap:", layerCanvas)

	points := []PointF{PointF{100, 100}, PointF{100, 200}, PointF{150, 280}, PointF{200, 300}, PointF{250, 200}, PointF{200, 130}, PointF{100, 100}}
	graphics.DrawCurve(pen, points)
	fmt.Println("DrawCurve.status:", graphics.LastResult)
	fmt.Println("DrawCurve.err:", graphics.LastError)

	//	graphics.DrawLines(pen, points)
	//	fmt.Println("DrawLines.status:", graphics.LastResult)
	//	fmt.Println("DrawLines.err:", graphics.LastError)

	// graphics

	BitBlt(hostDC, 0, 0, Width, Height, bufferDC, 0, 0, SRCCOPY)

	// d1 ULONG, d2, d3 WORD, d40, d41, d42, d43, d44, d45, d46, d47 BYTE
	// guid := NewGUID(1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1)
	guid := ImageFormatBMP
	encoderQuality := (uint)(50) //压缩比例
	var params EncoderParameters
	params.Count = 1
	params.Parameter[0].Guid = *EncoderQuality
	params.Parameter[0].NumberOfValues = 1
	params.Parameter[0].Type = PropertyTagTypeLong
	// uintptr(unsafe.Pointer(z))
	params.Parameter[0].Value = DATA_PTR(unsafe.Pointer(&encoderQuality))

	// ULONG encoderQuality = 50;                                //压缩比例
	// EncoderParameters encoderParameters;
	// encoderParameters.Count = 1;
	// encoderParameters.Parameter[0].Guid = EncoderQuality;
	// encoderParameters.Parameter[0].Type = EncoderParameterValueTypeLong;
	// encoderParameters.Parameter[0].NumberOfValues = 1;
	// encoderParameters.Parameter[0].Value = &encoderQuality;

	status := bitmapLayer.Save("c:\\zengzm\\gdi.bmp", (*CLSID)(guid), &params)
	fmt.Println("bitmap save status:", status)
	if graphics.LastError != nil {
		fmt.Println("graphics err:", graphics.LastError)
		return
	}

}
