package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/chai2010/tiff"
	"github.com/lukeroth/gdal"
)

func testGdal() {
	startTime := time.Now().UnixNano()

	var wg *sync.WaitGroup = new(sync.WaitGroup)

	for i := 0; i < 1; i++ {
		index := strconv.Itoa(i)
		wg.Add(1)
		go testOneTiff(index, wg)
	}

	wg.Wait()

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Println("time: ", seconds, "毫秒")
}

func drawTiff() {
	startTime := time.Now().UnixNano()

	filename := "C:\\temp\\A49C001003-0.tiff"
	dt, _ := gdal.Open(filename, 0)
	band := dt.RasterBand(1)
	band2 := dt.RasterBand(2)
	band3 := dt.RasterBand(3)
	sx := band.XSize()
	sy := band.YSize()
	bx, by := band.BlockSize()
	fmt.Println("size: ", sx, sy, bx, by)
	rdt := band.RasterDataType()
	fmt.Println("data type: ", rdt)
	// band.b
	// data := make([][]byte, sy)

	ix, iy := 1024, 768
	img := image.NewNRGBA(image.Rect(0, 0, ix, iy))
	rx := float64(ix) / float64(sx)
	ry := float64(iy) / float64(sy)
	fmt.Println("rest  ratio: ", rx, ry)

	for i := 0; i < sy; i++ {
		data := make([]uint8, sx)
		data2 := make([]uint8, sx)
		data3 := make([]uint8, sx)
		band.ReadBlock(0, i, unsafe.Pointer(&data[0]))
		band2.ReadBlock(0, i, unsafe.Pointer(&data2[0]))
		band3.ReadBlock(0, i, unsafe.Pointer(&data3[0]))
		for j := 0; j < sx; j++ {
			x, y := int(float64(i)*rx), int(float64(j)*ry)
			img.Set(x, y, color.RGBA{data[j], data2[j], data3[j], 255})
		}
	}
	imgfile, _ := os.Create("c:/temp/image.jpeg")
	jpeg.Encode(imgfile, img, nil)

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Println("time: ", seconds, "毫秒")
}

func testOneTiff(index string, wg *sync.WaitGroup) {
	defer wg.Done()

	filename := "C:\\temp\\A49C001003-" + index + ".tiff"
	fmt.Println("filename: ", filename)

	dt, _ := gdal.Open(filename, 0)
	// dt.BuildOverviews()
	fmt.Println("file list: ", dt.FileList())
	fmt.Println("GeoTransform: ", dt.GeoTransform())
	fmt.Println("Projection: ", dt.Projection())

	rc := dt.RasterCount()
	fmt.Println("Count: ", rc)

	band := dt.RasterBand(1)
	xs := band.XSize()
	ys := band.YSize()
	x, y := band.BlockSize()
	fmt.Println("size: ", xs, ys, x, y)
	// var data [10508 * 7028 * 1]byte

	// data := make([]byte, x*y*64)
	// gdal.Warp()
	rdt := band.RasterDataType()
	fmt.Println("data type: ", rdt)
	// band.b
	data := make([][]byte, 7028)

	sum := 0
	count := 0
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

	fmt.Println("sum of data: ", sum)
	fmt.Println("count: ", count)

	// band.ReadBlock(0,0,)
	fmt.Println("BandNumber: ", band.BandNumber())
	fmt.Println("OverviewCount: ", band.OverviewCount())
	for i := 0; i < band.OverviewCount(); i++ {
		ov := band.Overview(i)
		x, y := ov.BlockSize()
		fmt.Println("size: ", i, x, y)

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
