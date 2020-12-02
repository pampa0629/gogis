package main

import (
	"encoding/binary"
	"os"
)

func main() {
	testPath()
}

func testPath() {
	// p := ToAbsolutePath("c:/temp/a.b", "./c.d")
	// fmt.Println("path:", p)
}

func testFileWR() {
	filename := "c:/temp/abc"
	{
		f, _ := os.Create(filename)

		// data := []byte{'a', 'b', 'c'}
		str := "abc"
		f.WriteString(str)

		// binary.Write(f, binary.LittleEndian, data)
		f.Close()
	}

	{
		gix, _ := os.Open(filename)

		// var gixMark [3]byte
		gixMark := make([]byte, 3)
		// gix.Read(gixMark)
		binary.Read(gix, binary.LittleEndian, gixMark)
		gix.Close()
	}

}
