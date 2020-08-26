package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"io"
	"os"
	"strings"
	"sync"
	"unsafe"
)

// type shpPolyline struct {
// 	shpType                  // 图形类型，==3
// 	bbox       base.Rect2D    // 当前线状目标的坐标范围
// 	numParts  int32          // 当前线目标所包含的子线段的个数
// 	numPoints int32          // 当前线目标所包含的顶点个数
// 	parts     []int32        // 每个子线段的第一个坐标点在 Points 的位置
// 	points    []base.Point2D // 记录所有坐标点的数组
// }

// 枚举怎么弄呢？
type shpType int32

const (
	ShpNull     = 0
	ShpPoint    = 1
	ShpPolyLine = 3
	ShpPolygon  = 5
	// todo
	ShpMultiPoint  = 8
	ShpPointZ      = 11
	ShpPolyLineZ   = 13
	ShpPolygonZ    = 15
	ShpMultiPointZ = 18
	ShpPointM      = 21
	ShpPolyLineM   = 23
	ShpPolygonM    = 25
	ShpMultiPointM = 28
	ShpMultiPatch  = 31
)

type shpHeader struct {
	// big endian
	code    int32
	unuseds [5]int32
	length  int32
	// little endian
	version int32
	geoType shpType
	unused  int32 // 为了八字节对齐，增加4个字节；实际shape文件中并不存储此项内容
	xmin    float64
	ymin    float64
	xmax    float64
	ymax    float64
	zmin    float64
	zmax    float64
	mmin    float64
	mmax    float64
}

type shxRecordHeader struct {
	pos    int32 // big endian 单位：双字节
	length int32 // big endian 单位：双字节
}

type shpRecordHeadser struct {
	num    int32 // big endian
	length int32 // big endian 单位：双字节
}

type ShapeFile struct {
	Filename string
	shpHeader
	recordNum int               // 记录个数
	records   []shxRecordHeader // 记录头 数组
}

func readHeader(f *os.File, header *shpHeader) {
	p := (*[104]byte)(unsafe.Pointer(header)) // 对齐，所有多四个字节
	// n, err := f.Read((*p)[:]) 为啥两种写法都OK呢？
	_, err := f.Read(p[0:36])
	// fmt.Println("read shape file header num:", n)
	if err != nil && err != io.EOF {
		fmt.Println("read shape file header error", err)
	}

	header.code = base.ExEndian(header.code)
	if header.code != 9994 {
		fmt.Println("shape file code error")
	}
	header.length = base.ExEndian(header.length)

	// fmt.Printf("read shape file data:%x\n", p)

	f.Read(p[40:104])
	// fmt.Println("read shape file header:", header)
}

// 清空其它内存数据
func (this *ShapeFile) Close() {
	// shp.f.Close()

}

func (this *ShapeFile) Open(filename string) bool {
	// shp 文件
	// fmt.Println("file name length:", len(filename))
	this.Filename = filename
	shp, err := os.Open(filename)
	defer shp.Close()
	if err != nil {
		fmt.Println("open shape file error:", err)
	}

	// shp.f = bufio.NewReader(f)
	readHeader(shp, &this.shpHeader)

	// shx 文件
	shxName := strings.TrimSuffix(filename, ".shp") + ".shx"
	shx, _ := os.Open(shxName)
	defer shx.Close()

	var shxHeader shpHeader
	readHeader(shx, &shxHeader)
	info, _ := shx.Stat()
	// fmt.Println("shx file size:", info.Size())
	this.recordNum = (int)(info.Size()-100) / 8
	// fmt.Println("record num:", this.recordNum)

	// this.geometrys = make([]*shpPolyline, this.recordNum)

	// 读取shx中的记录头信息
	// recordNum int32             // 记录个数
	// records   []shxRecordHeader // 记录头 数组
	this.records = make([]shxRecordHeader, this.recordNum)
	// n, _ := shx.Read(ByteSlice(this.records))
	shx.Read(base.ByteSlice(this.records))
	for i := 0; i < this.recordNum; i++ {
		this.records[i].pos = base.ExEndian(this.records[i].pos)
		this.records[i].length = base.ExEndian(this.records[i].length)
	}
	// fmt.Println("read file length:", n)
	// fmt.Println("shp records:", shp.records)

	return true
}

// 加载start位置开始，批量读取geometry
func (this *ShapeFile) BatchLoad(start int, count int, features []Feature, wg *sync.WaitGroup) {
	// fmt.Println("ShapeFile.BatchLoad()", start, count, this.geoType, this.Filename)
	defer wg.Done()

	// 先确定要读取文件的位置和长度
	pos := uint64(this.records[start].pos) * 2
	len := int32(0)
	end := start + count
	for i := start; i < end; i++ {
		len += this.records[i].length*2 + 8 // 8个字节是 shp记录的头
	}

	f, err := os.Open(this.Filename)
	if err != nil {
		// return errors.New("open shape file error:" + err.Error())
		fmt.Println("open shape file error:", err.Error())
	}
	f.Seek(int64(pos), 0)
	data := make([]byte, len)
	f.Read(data)
	r := bytes.NewBuffer(data)
	// fmt.Println(len, this.Filename, f)
	// fmt.Println(features)

	for i := 0; i < count; i++ {
		features[i].geo = loadFromByte(r, this.geoType)
	}
	// return nil
}

// 从内存中读取一条记录
func loadFromByte(r io.Reader, shptype shpType) geometry.Geometry {
	// fmt.Println("ShapeFile.loadFromByte()")
	// 记录头
	var num, len int32
	binary.Read(r, binary.BigEndian, &num)
	binary.Read(r, binary.BigEndian, &len)

	var geo geometry.Geometry
	switch shptype {
	case ShpPoint:
		geo = loadShpOnePoint(r)
	case ShpPolyLine:
		geo = loadShpOnePolyline(r)
	case ShpPolygon:
		geo = loadShpOnePolygon(r)
	}
	return geo
}

func loadShpOnePoint(r io.Reader) geometry.Geometry {
	var geopoint geometry.GeoPoint
	var geotype shpType
	binary.Read(r, binary.LittleEndian, &geotype)
	binary.Read(r, binary.LittleEndian, &geopoint.X)
	binary.Read(r, binary.LittleEndian, &geopoint.Y)
	return &geopoint
}

func loadShpOnePolyline(r io.Reader) geometry.Geometry {
	var polyline geometry.GeoPolyline
	bbox, numParts, numPoints := loadShpOnePolyHeader(r)
	polyline.BBox = bbox

	parts := make([]int32, numParts+1)
	for i := int32(0); i < numParts; i++ {
		binary.Read(r, binary.LittleEndian, &parts[i])
	}
	parts[numParts] = numPoints

	polyline.Points = make([][]base.Point2D, numParts)
	for i := int32(0); i < numParts; i++ {
		polyline.Points[i] = make([]base.Point2D, parts[i+1]-parts[i])
		for j, _ := range polyline.Points[i] {
			binary.Read(r, binary.LittleEndian, &polyline.Points[i][j].X)
			binary.Read(r, binary.LittleEndian, &polyline.Points[i][j].Y)
		}
	}

	return &polyline
}

func loadShpOnePolygon(r io.Reader) geometry.Geometry {
	var polygon geometry.GeoPolygon
	bbox, numParts, numPoints := loadShpOnePolyHeader(r)
	polygon.BBox = bbox

	parts := make([]int32, numParts+1)
	for i := int32(0); i < numParts; i++ {
		binary.Read(r, binary.LittleEndian, &parts[i])
	}
	parts[numParts] = numPoints

	polygon.Points = make([][][]base.Point2D, numParts)
	for i := int32(0); i < numParts; i++ {
		polygon.Points[i] = make([][]base.Point2D, 1)
		polygon.Points[i][0] = make([]base.Point2D, parts[i+1]-parts[i])
		for j, _ := range polygon.Points[i][0] {
			binary.Read(r, binary.LittleEndian, &polygon.Points[i][0][j].X)
			binary.Read(r, binary.LittleEndian, &polygon.Points[i][0][j].Y)
		}
	}
	return &polygon
}

// 读取 polyline和polygon共同的部分
func loadShpOnePolyHeader(r io.Reader) (bbox base.Rect2D, numParts, numPoints int32) {
	var shptype int32
	binary.Read(r, binary.LittleEndian, &shptype)
	binary.Read(r, binary.LittleEndian, &bbox.Min.X)
	binary.Read(r, binary.LittleEndian, &bbox.Min.Y)
	binary.Read(r, binary.LittleEndian, &bbox.Max.X)
	binary.Read(r, binary.LittleEndian, &bbox.Max.Y)
	binary.Read(r, binary.LittleEndian, &numParts)
	binary.Read(r, binary.LittleEndian, &numPoints)
	return bbox, numParts, numPoints
}
