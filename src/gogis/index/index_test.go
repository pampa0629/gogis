package index

import (
	"bytes"
	"fmt"
	"gogis/base"
	"sort"
	"testing"
)

func TestCode(t *testing.T) {
	bits := []byte{1, 1, 1, 0}
	code := bits2code(bits)
	bits2 := Code2bits(code)
	if !bytes.Equal(bits2, bits) {
		t.Error("code2bits")
	}
	if bits2code(bits2) != code {
		t.Error("bits2code")
	}

}

func TestBits(t *testing.T) {
	// var index ZOrderIndex
	bits := []byte{0, 0, 1, 0}
	downBitss := buildDownBitss(bits)
	fmt.Println("bits:", bits, "downBitss:", downBitss)
}

func TestZOrder(t *testing.T) {
	var index ZOrderIndex
	ONE_CELL_COUNT = 5
	index.Init(base.NewRect2D(0, 0, 32, 32), 10000)
	index.String()
	if !bytes.Equal(index.AddOne2(base.NewRect2D(1, 1, 2, 2), 0), []byte{0, 0, 0, 0}) {
		t.Error("0")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(3, 1, 5, 2), 1), []byte{0, 0, 0, 0}) {
		t.Error("1")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(3, 3, 5, 5), 2), []byte{0, 0, 0, 0}) {
		t.Error("2")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(6, 3, 7, 5), 3), []byte{0, 0, 0, 0}) {
		t.Error("3")
	}

	if !bytes.Equal(index.AddOne2(base.NewRect2D(7, 3, 9, 5), 4), []byte{0, 0}) {
		t.Error("4")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(11, 3, 13, 6), 5), []byte{0, 0, 1, 0}) {
		t.Error("5")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(7, 7, 9, 9), 6), []byte{0, 0}) {
		t.Error("6")
	}

	if !bytes.Equal(index.AddOne2(base.NewRect2D(0, 16, 7.9, 24), 7), []byte{0, 1, 0, 0}) {
		t.Error("7")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(0, 24, 8, 32), 8), []byte{0, 1, 0, 1}) {
		t.Error("8")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(8, 16, 16, 24), 9), []byte{0, 1, 1, 0}) {
		t.Error("9")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(8, 24, 16, 32), 10), []byte{0, 1, 1, 1}) {
		t.Error("10")
	}

	if !bytes.Equal(index.AddOne2(base.NewRect2D(16, 16, 32, 32), 11), []byte{1, 1}) {
		t.Error("11")
	}

	if !bytes.Equal(index.AddOne2(base.NewRect2D(17, 13, 19, 16), 12), []byte{1, 0, 0, 1}) {
		t.Error("12")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(21, 13, 23, 16), 13), []byte{1, 0, 0, 1}) {
		t.Error("13")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(21, 9, 23, 11), 14), []byte{1, 0, 0, 1}) {
		t.Error("14")
	}
	if !bytes.Equal(index.AddOne2(base.NewRect2D(29, 13, 31, 15), 15), []byte{1, 0, 1, 1}) {
		t.Error("15")
	}

	// ===================================== //
	{
		ids := index.Query(base.NewRect2D(-1, -1, 33, 33))
		sort.Sort(base.Int64s(ids))
		fmt.Println("1 ids:", ids) // 0-15
	}
	// return

	{
		ids := index.Query(base.NewRect2D(0, 0, 32, 32))
		sort.Sort(base.Int64s(ids))
		fmt.Println("2 ids:", ids) // 0-11
	}

	{
		ids := index.Query(base.NewRect2D(7, 7, 25, 25))
		sort.Sort(base.Int64s(ids))
		fmt.Println("3 ids:", ids) // 6/7/8/9/10/11
	}

	{ //
		ids := index.Query(base.NewRect2D(8, 8, 24, 24))
		sort.Sort(base.Int64s(ids))
		fmt.Println("4 ids:", ids) // 6//9/11
	}

	{ //
		ids := index.Query(base.NewRect2D(9, 9, 23, 23))
		sort.Sort(base.Int64s(ids))
		fmt.Println("5 ids:", ids) // 6/9/11
	}

	{ // ？
		ids := index.Query(base.NewRect2D(8, 16, 16, 24))
		sort.Sort(base.Int64s(ids))
		fmt.Println("6 ids:", ids) // 9
	}

}
