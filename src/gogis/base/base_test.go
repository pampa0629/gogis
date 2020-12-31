package base

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestIO(t *testing.T) {
	fmt.Println(Float32ToBytes(11.1))
}

// 矩形与点
func TestRectPoint(t *testing.T) {
	var rect Rect2D
	rect.Min = Point2D{0, 0}
	rect.Max = Point2D{100, 100}

	// 矩形包含点
	if rect.IsContainsPoint(Point2D{0, 0}) {
		t.Errorf("矩形包含点错误1")
	}
	if !rect.IsContainsPoint(Point2D{50, 50}) {
		t.Errorf("矩形包含点错误2")
	}
	if rect.IsContainsPoint(Point2D{-1, -1}) {
		t.Errorf("矩形包含点错误3")
	}

	// 矩形覆盖点
	if !rect.IsCoverPoint(Point2D{0, 0}) {
		t.Errorf("矩形覆盖点错误1")
	}
	if !rect.IsCoverPoint(Point2D{50, 50}) {
		t.Errorf("矩形覆盖点错误2")
	}
	if rect.IsCoverPoint(Point2D{-1, -1}) {
		t.Errorf("矩形覆盖点错误3")
	}

}

// 矩形与矩形
func TestRectRect(t *testing.T) {
	var rect, rect1, rect2, rect3, rect4, rect5, rect6, rect7 Rect2D
	rect.Min = Point2D{0, 0}
	rect.Max = Point2D{100, 100}

	// 相离
	rect1.Min = Point2D{-100, -100}
	rect1.Max = Point2D{-10, -10}

	// 交集为点
	rect2.Min = Point2D{-50, -50}
	rect2.Max = Point2D{0, 0}

	// 交集为线
	rect3.Min = Point2D{-20, -20}
	rect3.Max = Point2D{0, 10}

	// 交集为面
	rect4.Min = Point2D{-20, -20}
	rect4.Max = Point2D{10, 10}

	// 交集为线和面，完全在内部
	rect5.Min = Point2D{0, 10}
	rect5.Max = Point2D{20, 20}

	// 交集为面，完全在内部
	rect6.Min = Point2D{10, 10}
	rect6.Max = Point2D{20, 20}

	// 完全一致
	rect7 = rect

	// 矩形包含矩形
	if rect.IsContains(rect1) {
		t.Errorf("矩形包含矩形错误1")
	}
	if rect.IsContains(rect2) {
		t.Errorf("矩形包含矩形错误2")
	}
	if rect.IsContains(rect3) {
		t.Errorf("矩形包含矩形错误3")
	}
	if rect.IsContains(rect4) {
		t.Errorf("矩形包含矩形错误4")
	}
	if rect.IsContains(rect5) {
		t.Errorf("矩形包含矩形错误5")
	}
	if !rect.IsContains(rect6) {
		t.Errorf("矩形包含矩形错误6")
	}
	if rect.IsContains(rect7) {
		t.Errorf("矩形包含矩形错误7")
	}

	// 矩形覆盖矩形
	if rect.IsCover(rect1) {
		t.Errorf("矩形覆盖矩形错误1")
	}
	if rect.IsCover(rect2) {
		t.Errorf("矩形覆盖矩形错误2")
	}
	if rect.IsCover(rect3) {
		t.Errorf("矩形覆盖矩形错误3")
	}
	if rect.IsCover(rect4) {
		t.Errorf("矩形覆盖矩形错误4")
	}
	if !rect.IsCover(rect5) {
		t.Errorf("矩形覆盖矩形错误5")
	}
	if !rect.IsCover(rect6) {
		t.Errorf("矩形覆盖矩形错误6")
	}
	if !rect.IsCover(rect7) {
		t.Errorf("矩形覆盖矩形错误7")
	}

	// 矩形与矩形相交
	if rect.IsIntersect(rect1) {
		t.Errorf("矩形相交矩形错误1")
	}
	if !rect.IsIntersect(rect2) {
		t.Errorf("矩形相交矩形错误2")
	}
	if !rect.IsIntersect(rect3) {
		t.Errorf("矩形相交矩形错误3")
	}
	if !rect.IsIntersect(rect4) {
		t.Errorf("矩形相交矩形错误4")
	}
	if !rect.IsIntersect(rect5) {
		t.Errorf("矩形相交矩形错误5")
	}
	if !rect.IsIntersect(rect6) {
		t.Errorf("矩形相交矩形错误6")
	}
	if !rect.IsIntersect(rect7) {
		t.Errorf("矩形相交矩形错误7")
	}

	// 矩形与矩形有交叠，即相交部分存在二维
	if rect.IsOverlap(rect1) {
		t.Errorf("矩形交叠矩形错误1")
	}
	if rect.IsOverlap(rect2) {
		t.Errorf("矩形交叠矩形错误2")
	}
	if rect.IsOverlap(rect3) {
		t.Errorf("矩形交叠矩形错误3")
	}
	if !rect.IsOverlap(rect4) {
		t.Errorf("矩形交叠矩形错误4")
	}
	if !rect.IsOverlap(rect5) {
		t.Errorf("矩形交叠矩形错误5")
	}
	if !rect.IsOverlap(rect6) {
		t.Errorf("矩形交叠矩形错误6")
	}
	if !rect.IsOverlap(rect7) {
		t.Errorf("矩形交叠矩形错误7")
	}

}

func TestPath(t *testing.T) {
	{
		p1 := "c:/temp/cache"
		p2 := "c:/temp/cache/"
		p := "mapname"
		res1 := filepath.Join(p1, p)
		res1 = filepath.ToSlash(res1)
		if res1 != "c:/temp/cache/mapname" {
			t.Errorf("连接文件夹错误1")
		}
		res2 := filepath.Join(p2, p)
		res2 = filepath.ToSlash(res2)
		if res2 != "c:/temp/cache/mapname" {
			t.Errorf("连接文件夹错误1")
		}
	}

	{
		p1 := "c:/temp/cache/abc.ext"
		p2 := "c:\\temp\\cache\\abc.ext"
		p3 := "/temp/cache/abc.ext"
		title := "abc"
		if GetTitle(p1) != title {
			t.Errorf("得到文件title错误1")
		}
		if GetTitle(p2) != title {
			t.Errorf("得到文件title错误2")
		}
		if GetTitle(p3) != title {
			t.Errorf("得到文件title错误3")
		}
	}

	{
		p1 := "c:/user/temp/a.b"
		r := "./c.d"
		r2 := "c.d"
		p2 := "c:/user/temp/c.d"

		if GetRelativePath(p1, p2) != r2 {
			t.Errorf("得到相对路径错误1")
		}

		a := GetAbsolutePath(p1, r)
		if a != p2 {
			t.Errorf("得到绝对路径错误1")
		}

		a2 := GetAbsolutePath(p1, r2)
		if a2 != p2 {
			t.Errorf("得到绝对路径错误12")
		}
	}

	{
		p1 := "/user/temp/a.b"
		r := "./c.d"
		r2 := "c.d"
		p2 := "/user/temp/c.d"

		if GetRelativePath(p1, p2) != r2 {
			t.Errorf("得到相对路径错误2")
		}

		a := GetAbsolutePath(p1, r)
		if a != p2 {
			t.Errorf("得到绝对路径错误2")
		}

		a2 := GetAbsolutePath(p1, r2)
		if a2 != p2 {
			t.Errorf("得到绝对路径错误22")
		}
	}

}
