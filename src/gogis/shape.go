package gogis

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unsafe"
)

type shpPoint Point2D

// Box 记录了当前的线目标的坐标范围，它是一个 double 型的数组，按照 Xmin 、 Ymin 、 Xmax 、 Ymax 的顺序记录了坐标范围；
type shpBox Rect2D

// type shpBox struct {
// 	xmin float64
// 	ymin float64
// 	xmax float64
// 	ymax float64
// }

type shpPolyline struct {
	shpType              // 图形类型，==3
	box       Rect2D     // 当前线状目标的坐标范围
	numParts  int32      // 当前线目标所包含的子线段的个数
	numPoints int32      // 当前线目标所包含的顶点个数
	parts     []int32    // 每个子线段的第一个坐标点在 Points 的位置
	points    []shpPoint // 记录所有坐标点的数组
}

// 枚举怎么弄呢？
type shpType int32

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
	pos    int32 // big endian
	length int32 // big endian
}

type shpRecordHeadser struct {
	num    int32 // big endian
	length int32 // big endian
}

type ShapeFile struct {
	Filename string
	shpHeader
	// f         *bufio.Reader     // shp file handle
	// f         *os.File
	recordNum int               // 记录个数
	records   []shxRecordHeader // 记录头 数组
	geometrys []*shpPolyline    // 几何对象的数组
	index     *GridIndex        // 空间索引
}

func readHeader(f *os.File, header *shpHeader) {
	p := (*[104]byte)(unsafe.Pointer(header)) // 对齐，所有多四个字节
	// n, err := f.Read((*p)[:]) 为啥两种写法都OK呢？
	n, err := f.Read(p[0:36])
	fmt.Println("read shape file header num:", n)
	if err != nil && err != io.EOF {
		fmt.Println("read shape file header error", err)
	}

	header.code = exEndian(header.code)
	if header.code != 9994 {
		fmt.Println("shape file code error")
	}
	header.length = exEndian(header.length)

	// fmt.Printf("read shape file data:%x\n", p)

	n, _ = f.Read(p[40:104])
	fmt.Println("read shape file header:", header)
}

// 清空其它内存数据
func (this *ShapeFile) Close() {
	// shp.f.Close()

}

func (this *ShapeFile) Open(filename string) bool {
	// shp 文件
	fmt.Println("file name length:", len(filename))
	this.Filename = filename
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("open shape file error:", err)
	}

	// shp.f = bufio.NewReader(f)
	readHeader(f, &this.shpHeader)

	// shx 文件
	shxName := strings.TrimSuffix(filename, ".shp") + ".shx"
	fmt.Println("shx  file name:", shxName)
	shx, _ := os.Open(shxName)
	defer shx.Close()

	var shxHeader shpHeader
	readHeader(shx, &shxHeader)
	info, _ := shx.Stat()
	fmt.Println("shx file size:", info.Size())
	this.recordNum = (int)(info.Size()-100) / 8
	fmt.Println("record num:", this.recordNum)

	this.geometrys = make([]*shpPolyline, this.recordNum)

	// 读取shx中的记录头信息
	// recordNum int32             // 记录个数
	// records   []shxRecordHeader // 记录头 数组
	this.records = make([]shxRecordHeader, this.recordNum)
	n, _ := shx.Read(ByteSlice(this.records))
	for i := 0; i < this.recordNum; i++ {
		this.records[i].pos = exEndian(this.records[i].pos)
		this.records[i].length = exEndian(this.records[i].length)
	}
	fmt.Println("read file length:", n)
	// fmt.Println("shp records:", shp.records)

	return true
}

// 一次性从文件加载到内存的记录个数
const ONE_LOAD_COUNT = 100000

type ShpRead struct {
	pos int
	len int
}

// 用多文件读取的方式，把geometry都转载到内存中
func (this *ShapeFile) Load() {
	// 计算一下，需要加载多少次
	forcount := (int)(this.recordNum/ONE_LOAD_COUNT) + 1
	fmt.Println("forcount:", forcount)
	// datas := make([][]byte, forcount)

	reads := make([]ShpRead, forcount)

	for i := 0; i < forcount; i++ {
		num := i * ONE_LOAD_COUNT
		// N条记录打包一次性读取
		len := 0
		for i := 0; i < ONE_LOAD_COUNT && num < this.recordNum; i++ {
			len += (int)(this.records[num].length*2 + 8)
			num++
		}
		if i == 0 {
			reads[0].pos = 100
		} else {
			reads[i].pos = reads[i-1].pos + reads[i-1].len
		}

		reads[i].len = len
	}
	this.Close()
	// fmt.Println("reads:", reads)

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < forcount; i++ {
		wg.Add(1)
		go this.loadBatch(i*ONE_LOAD_COUNT, reads[i], wg)
	}
	wg.Wait()
	this.BuildSpatialIndex()
}

// 从num位置开始，批量读取geometry
func (this *ShapeFile) loadBatch(num int, read ShpRead, wg *sync.WaitGroup) {
	pos := (uint64)(0)
	f, err := os.Open(this.Filename)
	if err != nil {
		fmt.Println("open shape file error:", err)
	}
	f.Seek(int64(read.pos), 0)
	data := make([]byte, read.len)
	f.Read(data)

	for i := 0; i < ONE_LOAD_COUNT && num < this.recordNum; i++ {
		this.geometrys[num], pos = loadFromByte(data, pos)
		num++
	}
	defer wg.Done()
}

// 从内存中读取一条记录
func loadFromByte(data []byte, pos uint64) (*shpPolyline, uint64) {

	// type shpPolyline struct {
	// 	shpType              // 图形类型，==3
	// 	box       shpBox     // 当前线状目标的坐标范围
	// 	numParts  int32      // 当前线目标所包含的子线段的个数
	// 	numPoints int32      // 当前线目标所包含的顶点个数
	// 	parts     []int32    // 每个子线段的第一个坐标点在 Points 的位置
	// 	points    []shpPoint // 记录所有坐标点的数组
	// }
	// 记录头
	var recordHeader shpRecordHeadser
	recordHeader.num = toInt32(data, pos, true)
	pos += 4
	recordHeader.length = toInt32(data, pos, true)
	pos += 4
	// fmt.Println("shp one record header:", recordHeader)

	// 记录体
	var polyline shpPolyline
	polyline.shpType = (shpType)(toInt32(data, pos, false))
	pos += 4
	polyline.box.Min.X = toFloat64(data, pos)
	pos += 8
	polyline.box.Min.Y = toFloat64(data, pos)
	pos += 8
	polyline.box.Max.X = toFloat64(data, pos)
	pos += 8
	polyline.box.Max.Y = toFloat64(data, pos)
	pos += 8
	polyline.numParts = toInt32(data, pos, false)
	pos += 4
	polyline.numPoints = toInt32(data, pos, false)
	pos += 4

	// parts
	polyline.parts = make([]int32, polyline.numParts)
	copy(ByteSlice(polyline.parts), data[pos:])
	pos += (uint64)(4 * polyline.numParts)

	// points
	polyline.points = make([]shpPoint, polyline.numPoints)
	copy(ByteSlice(polyline.points), data[pos:])
	pos += (uint64)(16 * polyline.numPoints)

	// fmt.Println("shp one record:", polyline)
	// fmt.Printf("shp one data:%x\n", data[0:])
	return &polyline, pos
}

func (this *ShapeFile) BuildSpatialIndex() {
	if this.index == nil {
		this.index = new(GridIndex)
		var bbox Rect2D
		bbox.Min.X = this.xmin
		bbox.Min.Y = this.ymin
		bbox.Max.X = this.xmax
		bbox.Max.Y = this.ymax
		this.index.Init(bbox, this.recordNum)
		// fmt.Println("shp grid index:", this.index)
		this.index.Build(this.geometrys)
	}
}

// 计算索引重复度，为后续有可能增加多级格网做准备
func (this *ShapeFile) calcRepeatability() float64 {
	indexCount := 0.0
	for i := 0; i < this.index.row; i++ {
		for j := 0; j < this.index.col; j++ {
			indexCount += float64(len(this.index.indexs[i][j]))
		}
	}
	repeat := indexCount / float64(this.recordNum)
	fmt.Println("shp index重复度为:", repeat)
	return repeat
}

// 根据空间范围查询，返回范围内geo的ids
func (this *ShapeFile) Query(bbox Rect2D) []int {
	ids := make([]int, 0)
	minRow, maxRow, minCol, maxCol := this.index.GetGridNo(bbox)
	fmt.Println("shp file query:", minRow, maxRow, minCol, maxCol)

	// 最后赋值
	for i := minRow; i <= maxRow; i++ { // 高度（y方向）代表行
		for j := minCol; j <= maxCol; j++ {
			// fmt.Println("this.index.indexs:", i, j, this.index.indexs[i][j])
			ids = append(ids, this.index.indexs[i][j]...)
		}
	}

	// 这里应该还要去掉重复id todo ......
	return ids
}
