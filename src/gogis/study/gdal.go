package main

import (
	"bytes"
	"fmt"
	"gogis/base"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/chai2010/tiff"
	"github.com/lukeroth/gdal"
)

func gdalmain() {
	// testGdal()
	// drawTiffs()
	drawTiff(0, nil)
	// openList()
}

func openList() {
	filename := "C:/BigData/10_Data/testimage/image/filelist.txt"
	data, _ := ioutil.ReadFile(filename)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		fmt.Println(line)
		dt, _ := gdal.Open(line, 0)
		fmt.Println(dt)
	}
}

func testGdal() {
	tr := base.NewTimeRecorder()

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	count := 100

	for i := 0; i < count; i++ {
		index := strconv.Itoa(i)
		wg.Add(1)
		go testOneTiff(index, wg)
	}

	wg.Wait()

	tr.Output(strconv.Itoa(count))
}

func drawTiffs() {
	tr := base.NewTimeRecorder()
	tr.Output("main begin")
	// var wg *sync.WaitGroup = new(sync.WaitGroup)
	gm := new(base.GoMax)
	gm.Init(10)
	count := 20

	for i := 0; i < count; i++ {
		gm.Add()
		go drawTiff(i, gm)
	}

	gm.Wait()
	tr.Output(strconv.Itoa(count))
}

func drawTiff(i int, gm *base.GoMax) {
	if gm != nil {
		defer gm.Done()
	}

	tr := base.NewTimeRecorder()
	str := strconv.Itoa(i)

	// filename := "C:\\temp\\A49C001003-0.tiff"
	// filename := "C:\\temp\\A49C001003-" + strconv.Itoa(i%5) + ".tif"
	filename := "C:\\BigData\\10_Data\\testimage\\image\\C49C001003.tif"
	dt, _ := gdal.Open(filename, 0)
	fmt.Println("Projection:", dt.Projection())
	fmt.Println("GeoTransform:", dt.GeoTransform())

	// bc := dt.RasterCount()
	// ovCount := 2
	// ovList := []int{4, 16}
	// fun := func(complete float64, message string, progressArg interface{}) int {
	// 	fmt.Println(complete, message)
	// 	return 1
	// }
	// err := dt.BuildOverviews("nearest", ovCount, ovList, bc, []int{1, 2, 3}, fun, nil)
	// base.PrintError("BuildOverviews", err)
	// fmt.Println(dt.Metadata)

	band1 := dt.RasterBand(1) // band 从1 起
	band2 := dt.RasterBand(2)
	band3 := dt.RasterBand(3)
	b1 := band1
	ovCount := band1.OverviewCount()
	fmt.Println("OverviewCount: ", ovCount)
	for i := -1; i < ovCount; i++ {
		fmt.Println("XY Size: ", i, b1.XSize(), b1.YSize())
		s1, s2 := b1.BlockSize()
		fmt.Println("BlockSize: ", i, s1, s2)
		b1 = band1.Overview(i + 1) // overview从0起
	}

	sx := band1.XSize()
	sy := band1.YSize()

	ix, iy := 1024, 768
	img := image.NewRGBA(image.Rect(0, 0, ix, iy))
	rx := float64(ix) / float64(sx)
	ry := float64(iy) / float64(sy)
	// fmt.Println("rect  ratio: ", rx, ry)

	data := make([]uint8, ix*4)

	for i := 0; i < sy; i++ {
		data1 := make([]uint8, sx)
		data2 := make([]uint8, sx)
		data3 := make([]uint8, sx)
		band1.ReadBlock(0, i, unsafe.Pointer(&data1[0]))
		band2.ReadBlock(0, i, unsafe.Pointer(&data2[0]))
		band3.ReadBlock(0, i, unsafe.Pointer(&data3[0]))
		for j := 0; j < sx; j++ {
			x := int(float64(j) * rx)
			dx := 4 * x
			data[dx+0] = data1[j]
			data[dx+1] = data2[j]
			data[dx+2] = data3[j]
			data[dx+3] = 255
		}
		y := int(float64(i) * ry)
		copy(img.Pix[img.Stride*y:], data)
	}
	imgfile, _ := os.Create("c:/temp/image" + str + ".jpeg")
	jpeg.Encode(imgfile, img, nil)

	tr.Output(str + " draw")
}

func testOneTiff(index string, wg *sync.WaitGroup) {
	defer wg.Done()

	// filename := "C:\\temp\\A49C001003-" + index + ".tiff"
	filename := "C:\\temp\\A49C001003.tiff"
	// fmt.Println("filename: ", filename)

	dt, _ := gdal.Open(filename, 0)
	// dt.BuildOverviews()
	// fmt.Println("file list: ", dt.FileList())
	// fmt.Println("GeoTransform: ", dt.GeoTransform())
	// fmt.Println("Projection: ", dt.Projection())

	// rc := dt.RasterCount()
	// fmt.Println("Count: ", rc)

	band := dt.RasterBand(1)
	// xs := band.XSize()
	ys := band.YSize()
	// x, y := band.BlockSize()
	// fmt.Println("size: ", xs, ys, x, y)
	// var data [10508 * 7028 * 1]byte

	// data := make([]byte, x*y*64)
	// gdal.Warp()
	// rdt := band.RasterDataType()
	// fmt.Println("data type: ", rdt)
	// band.b
	data := make([][]byte, 7028)

	// sum := 0
	// count := 0
	for i := 0; i < ys; i++ {
		// go func(band gdal.RasterBand, data [][]byte, i int) {
		data[i] = make([]byte, 10508)
		band.ReadBlock(0, i, unsafe.Pointer(&data[i][0]))
		// for _, v := range data[i] {
		// 	sum += int(v)
		// if v != 0 {
		// 	// fmt.Println("value: ", i, v)
		// 	count++
		// }
		// }
		// }(band, data, i)
	}

	// fmt.Println("sum of data: ", sum)
	// fmt.Println("count: ", count)

	// band.ReadBlock(0,0,)
	// fmt.Println("BandNumber: ", band.BandNumber())
	// fmt.Println("OverviewCount: ", band.OverviewCount())
	for i := 0; i < band.OverviewCount(); i++ {
		// ov := band.Overview(i)
		// x, y := ov.BlockSize()
		// fmt.Println("size: ", i, x, y)

	}

}

func testTiff() {
	tiffname := "C:\\temp\\A49C001003.tiff"

	data, _ := ioutil.ReadFile(tiffname)
	fmt.Println(len(data))

	// Decode tiff
	img, err := tiff.Decode(bytes.NewReader(data))

	fmt.Println(err)
	fmt.Println(img.Bounds(), img.ColorModel())
	// fmt.Println(img)
	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()
	// add := uint32(0)
	for i := 0; i < dy; i++ {
		for j := 0; j < dx; j++ {
			clr := img.At(i, j)
			r, g, b, a := clr.RGBA()
			one := r + g + b
			if one != 0 {
				fmt.Println(i, j, one, clr, r, g, b, a)
			}
			// add += r
			// add += g
			// add += b
			// add += a
		}
	}

}
