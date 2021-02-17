package index

import (
	"fmt"
	"gogis/base"
	"reflect"
	"testing"
)

func TestCode(t *testing.T) {
	// bits := []byte{1, 1, 1, 0}
	// code := Bits2code(bits)
	// bits2 := Code2bits(code)
	// if !bytes.Equal(bits2, bits) {
	// 	t.Error("code2bits")
	// }
	// if Bits2code(bits2) != code {
	// 	t.Error("bits2code")
	// }

}

func TestBits(t *testing.T) {
	// var index ZOrderIndex
	bits := []byte{0, 0, 1, 0}
	downBitss := buildDownBitss(bits)
	fmt.Println("bits:", bits, "downBitss:", downBitss)
}

func TestXzorderCode(t *testing.T) {
	var idx XzorderIndex
	var bbox base.Rect2D
	bbox.Max.X, bbox.Max.Y = 32, 32
	idx.InitDB(bbox, 2)
	{
		var bbox base.Rect2D
		bbox.Min.X, bbox.Min.Y = 15, 15
		bbox.Max.X, bbox.Max.Y = 17, 17
		code := idx.GetCode(bbox)
		if code != 8 {
			t.Error("get code 1")
		}
	}
	{
		var bbox base.Rect2D
		bbox.Min.X, bbox.Min.Y = 3, 3
		bbox.Max.X, bbox.Max.Y = 5, 5
		code := idx.GetCode(bbox)
		if code != 5 {
			t.Error("get code 2")
		}
	}
	{
		idx.level = 2
		var bbox base.Rect2D
		bbox.Min.X, bbox.Min.Y = 15, 15
		bbox.Max.X, bbox.Max.Y = 25, 55
		code := idx.GetCode(bbox)
		if code != 1 {
			t.Error("get code 3")
		}
	}
}

func TestXzorderQuery(t *testing.T) {
	var idx XzorderIndex
	var bbox base.Rect2D
	bbox.Max.X, bbox.Max.Y = 32, 32
	idx.InitDB(bbox, 2)
	{
		var bbox base.Rect2D
		bbox.Min.X, bbox.Min.Y = 7, 7
		bbox.Max.X, bbox.Max.Y = 17, 17
		codes := idx.QueryDB(bbox)
		if !reflect.DeepEqual(codes, []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 13, 14, 17}) {
			t.Error("query codes 5")
		}
	}
	{
		var bbox base.Rect2D
		bbox.Min.X, bbox.Min.Y = -1, -1
		bbox.Max.X, bbox.Max.Y = 1, 1
		codes := idx.QueryDB(bbox)
		if !reflect.DeepEqual(codes, []int32{0, 1, 5}) {
			t.Error("query codes 4")
		}
	}
	{
		var bbox base.Rect2D
		bbox.Min.X, bbox.Min.Y = -1, -1
		bbox.Max.X, bbox.Max.Y = 33, 33
		codes := idx.QueryDB(bbox)
		if !reflect.DeepEqual(codes, []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}) {
			t.Error("query codes 3")
		}
	}
	{
		var bbox base.Rect2D
		bbox.Min.X, bbox.Min.Y = 15, 15
		bbox.Max.X, bbox.Max.Y = 17, 17
		codes := idx.QueryDB(bbox)
		if !reflect.DeepEqual(codes, []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 13, 14, 17}) {
			t.Error("query codes 1")
		}
	}
	{
		var bbox base.Rect2D
		bbox.Min.X, bbox.Min.Y = 17, 17
		bbox.Max.X, bbox.Max.Y = 18, 18
		codes := idx.QueryDB(bbox)
		if !reflect.DeepEqual(codes, []int32{0, 1, 2, 3, 4, 8, 11, 14, 17}) {
			t.Error("query codes 2")
		}
	}
}
