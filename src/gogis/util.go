package gogis

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"reflect"
	"unsafe"
)

type Point2D struct {
	X float64
	Y float64
}

type Rect2D struct {
	Min, Max Point2D
	// xmin float64
	// ymin float64
	// xmax float64
	// ymax float64
}

// 初始化，使之无效，即min为浮点数最大值；max为浮点数最小值。而非均为0
func (this *Rect2D) Init() {
	this.Min.X = math.MaxFloat64
	this.Min.Y = math.MaxFloat64
	this.Max.X = -math.MaxFloat64
	this.Max.Y = -math.MaxFloat64
}

func NewRect2D(minx, miny, maxx, maxy float64) (value Rect2D) {
	value.Min.X = minx
	value.Min.Y = miny
	value.Max.X = maxx
	value.Max.Y = maxy
	return
}

// 两个box合并，取并集的box
func (this *Rect2D) Union(rect Rect2D) {
	this.Min.X = math.Min(this.Min.X, rect.Min.X)
	this.Min.Y = math.Min(this.Min.Y, rect.Min.Y)
	this.Max.X = math.Max(this.Max.X, rect.Max.X)
	this.Max.Y = math.Max(this.Max.Y, rect.Max.Y)
}

// 计算面积
func (this *Rect2D) Area() float64 {
	return this.Dx() * this.Dy()
}

func (this *Rect2D) Dx() float64 {
	return this.Max.X - this.Min.X
}

func (this *Rect2D) Dy() float64 {
	return this.Max.Y - this.Min.Y
}

func IntMin(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func IntMax(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func deleteEmptyDir(path string) {
	dir, _ := ioutil.ReadDir(path)
	if len(dir) == 0 {
		os.RemoveAll(path)
	}
}

// 根据bbox和对象数量，计算缓存的最小最大合适层级
// 再小的层级没有必要（图片上的显示范围太小）；再大的层级则瓦片上对象太稀疏
func calcMinMaxLevels(bbox Rect2D, geoCount int) (minLevel, maxLevel int) {
	fmt.Println("bbox:", bbox)
	minLevel = 0
	dis := 180.0
	dx := bbox.Dx()
	dy := bbox.Dy()
	// 地图长宽加一起，得够一个瓦片的长度，才值得从这个层级出图；不然就继续放大层级
	for dx+dy < dis {
		minLevel++
		dis /= 2.0
	}

	// 最大层级计算，要求每个瓦片的平均对象个数不少于 特定个数
	const GEO_MIN_COUNT = 10 // 考虑空白区域
	maxLevel = minLevel      // 从最小层级开始，后面每次瓦片数量*4
	for geoCount > GEO_MIN_COUNT {
		maxLevel++
		geoCount /= 4
	}

	return
}

// 计算特定层级，每个像素的经纬度长度；按照每个瓦片256*256计算
func calcLevelDis(level int) float64 {
	dis := 360.0
	for level > 0 {
		level--
		dis /= 2.0
	}
	return dis / 256.0
}

// 计算两点距离的平方
func distanceSquare(x0, y0, x1, y1 float64) float64 {
	return math.Pow((x0-x1), 2) + math.Pow((y0-y1), 2)
}

// 得到文件名中的 title部分；输入：C:/temp/JBNTBHTB.shp ，返回 JBNTBHTB
func GetTile(fullname string) string {
	filenameall := path.Base(fullname)
	filesuffix := path.Ext(fullname)
	fileprefix := filenameall[0 : len(filenameall)-len(filesuffix)]
	return fileprefix
}

// 把任意类型的切片转换为byte[]，内存地址一致，以便数据读写拷贝等用途
func ByteSlice(slice interface{}) (data []byte) {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Slice {
		panic(fmt.Sprintf("ByteSlice called with non-slice value of type %T", slice))
	}
	h := (*reflect.SliceHeader)((unsafe.Pointer(&data)))
	h.Cap = sv.Cap() * int(sv.Type().Elem().Size())
	h.Len = sv.Len() * int(sv.Type().Elem().Size())
	h.Data = sv.Pointer()
	return
}

// 把任意结构转换为byte[]，内存地址一致，以便数据读写拷贝等用途
// 测试不过，暂时封存
func xxByteStruct(value interface{}, size int) (data []byte) {
	sv := reflect.ValueOf(value)
	fmt.Println("struct kind is:", sv.Kind())
	if sv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("ByteStruct called with non-struct value of type %T", value))
	}
	h := (*reflect.SliceHeader)((unsafe.Pointer(&data)))
	// h.Cap = sv.Cap() * int(sv.Type().Elem().Size())
	h.Cap = size //  (int)(unsafe.Sizeof(value))
	h.Len = size //  (int)(unsafe.Sizeof(value))
	h.Data = (uintptr)(unsafe.Pointer(&value))
	fmt.Println("struct len is:", h.Len)
	return
}

// 大小端互换
func exEndian(value int32) int32 {
	buf := (*[4]byte)(unsafe.Pointer(&value))
	buf[0], buf[3] = buf[3], buf[0]
	buf[1], buf[2] = buf[2], buf[1]
	return value
}

// 大小端互换
func exEndianDouble(value float64) float64 {
	buf := (*[8]byte)(unsafe.Pointer(&value))
	buf[0], buf[7] = buf[7], buf[0]
	buf[1], buf[6] = buf[6], buf[1]
	buf[2], buf[5] = buf[5], buf[2]
	buf[3], buf[4] = buf[4], buf[3]
	return value
}

// 从切片中读取数据
func toInt32(input []byte, pos uint64, change bool) (data int32) {
	bytesBuffer := bytes.NewBuffer(input[pos : pos+4])
	if change {
		binary.Read(bytesBuffer, binary.BigEndian, &data)
	} else {
		binary.Read(bytesBuffer, binary.LittleEndian, &data)
	}
	return
}

// 从切片中读取数据
func toFloat64(input []byte, pos uint64) (data float64) {
	bytesBuffer := bytes.NewBuffer(input[pos : pos+8])
	binary.Read(bytesBuffer, binary.LittleEndian, &data)
	return
}

// 从文件中读取数据
func readInt32(f *os.File, change bool) (data int32) {
	p := (*[4]byte)(unsafe.Pointer(&data))
	// n, err := f.Read((*p)[:]) 为啥两种写法都OK呢？
	_, _ = f.Read(p[:])
	if change {
		data = exEndian(data)
	}
	return
}

// 从文件中读取数据
func readFloat64(f *os.File) (data float64) {
	p := (*[8]byte)(unsafe.Pointer(&data))
	// n, err := f.Read((*p)[:]) 为啥两种写法都OK呢？
	_, _ = f.Read(p[:])
	return
}

// 计算并设置 web出图合适的 绘制参数params
// todo
func SetParams(gmap *Map, nmap *Map, size int, row int, col int) {
	// 根据 row  和 col 修改 map的bbox
	// dx := gmap.BBox.Dx() / 4
	// dy := gmap.BBox.Dy() / 4
	// todo 1024 的修改
	change := 1024 / float64(size)
	scale := nmap.canvas.params.scale
	dx := float64(gmap.canvas.params.dx) / scale / change
	dy := float64(gmap.canvas.params.dy) / scale / change

	nmap.BBox = gmap.BBox
	nmap.BBox.Min.X += float64(col) * dx
	nmap.BBox.Max.X = nmap.BBox.Min.X + dx

	nmap.BBox.Max.Y -= float64(row) * dy
	nmap.BBox.Min.Y = nmap.BBox.Max.Y - dy
}
