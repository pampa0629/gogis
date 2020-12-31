package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LindsayBradford/go-dbf/godbf"
)

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
	Code    int32
	Unuseds [5]int32
	Length  int32
	// little endian
	Version int32
	GeoType geometry.ShpType
	Xmin    float64
	Ymin    float64
	Xmax    float64
	Ymax    float64
	Zmin    float64
	Zmax    float64
	Mmin    float64
	Mmax    float64
}

func (this *shpHeader) read(r io.Reader) {
	binary.Read(r, binary.LittleEndian, this)
	this.Code = base.ExEndian(this.Code)
	this.Length = base.ExEndian(this.Length)

	if this.Code != 9994 {
		fmt.Println("shape file code error")
	}
	// fmt.Println("shp header:", this)
}

type shxRecordHeader struct {
	// 注意：pos和length的单位都是双字节，需要*2后才能 seek 文件位置
	Pos    int32 // big endian 单位：双字节
	Length int32 // big endian 单位：双字节
}

type ShapeFile struct {
	Filename  string            //  shp主文件名
	shpHeader                   // 文件头
	recordNum int               // 记录个数
	records   []shxRecordHeader // 记录头 数组
	table     *godbf.DbfTable
}

// dbf 字段类型
// const (
// 	Character DbaseDataType = 'C'
// 	Logical   DbaseDataType = 'L'
// 	Date      DbaseDataType = 'D'
// 	Numeric   DbaseDataType = 'N'
// 	Float     DbaseDataType = 'F'
// )

// 清空内存
func (this *ShapeFile) Close() {
	this.records = this.records[:0]
	// this.table.Close()
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
	this.shpHeader.read(shp)

	// shx 文件
	shxName := strings.TrimSuffix(filename, ".shp") + ".shx"
	data, _ := ioutil.ReadFile(shxName)
	// 读取shx中的记录头信息
	this.recordNum = (int)(len(data)-100) / 8
	this.records = make([]shxRecordHeader, this.recordNum)
	r := bytes.NewReader(data[100:])
	binary.Read(r, binary.BigEndian, this.records)

	// dbf 文件
	// dbfName := strings.TrimSuffix(filename, ".shp") + ".dbf"
	// this.table, _ = godbf.NewFromFile(dbfName, "UTF8")

	return true
}

func (this *ShapeFile) GetFieldInfos() (finfos []FieldInfo) {
	fds := this.table.Fields()
	finfos = make([]FieldInfo, len(fds))
	for i, ds := range fds {
		finfos[i].Name = ds.Name()
		finfos[i].Length = int(ds.Length())
		finfos[i].Type = dbfTypeConvertor(ds.FieldType())
	}
	return
}

func dbfTypeConvertor(dtype godbf.DbaseDataType) FieldType {
	ftype := TypeUnknown
	switch dtype {
	case godbf.Character:
		ftype = TypeString
	case godbf.Logical:
		ftype = TypeBool
	case godbf.Date:
		ftype = TypeTime
	case godbf.Numeric: // 字符串数字，也可以是浮点数
		ftype = TypeFloat
	case godbf.Float:
		ftype = TypeFloat
	}
	return ftype
}

// 读取一条记录
func (this *ShapeFile) LoadOne(f *os.File, id int) (feature Feature) {

	// 先确定要读取文件的位置和长度
	pos := uint64(this.records[id].Pos) * 2
	// length := this.records[id].Length*2 + 8

	// f, err := os.Open(this.Filename)
	// if err != nil {
	// 	fmt.Println("open shape file error:", err.Error())
	// }
	f.Seek(int64(pos), 0)

	feature.Geo = loadFromByte(f, this.GeoType)
	return

	// 属性字段处理
	// fieldInfos := this.table.Fields()
	// fields := this.table.FieldNames()
	// fieldCount := len(fieldInfos)
	// for i := 0; i < count; i++ {

	// features[i].Atts = make(map[string]interface{}, fieldCount)
	// for j, name := range fields {
	// 	features[i].Atts[name] = dbfString2Value(this.table.FieldValue(i+start, j), fieldInfos[j].FieldType())
	// }
	// }
}

// 加载start位置开始，批量读取geometry
// 要求读取的count都是连续存储的
func (this *ShapeFile) BatchLoad(f *os.File, start int, count int, features []Feature, wg *sync.WaitGroup) {
	// fmt.Println("ShapeFile.BatchLoad()", start, count, this.geoType, this.Filename)
	if wg != nil {
		defer wg.Done()
	}

	// 先确定要读取文件的位置和长度
	pos := uint64(this.records[start].Pos) * 2
	length := int32(0)
	end := start + count
	for i := start; i < end; i++ {
		length += this.records[i].Length*2 + 8 // 8个字节是 shp记录的头
	}

	if f == nil {
		f, _ = os.Open(this.Filename)
		defer f.Close()
		// if err != nil {
		// 	fmt.Println("open shape file error:", err.Error())
		// }
	}

	f.Seek(int64(pos), 0)
	data := make([]byte, length)
	f.Read(data)
	// fmt.Println("read shp  count:", n, " error:", err)
	r := bytes.NewBuffer(data)

	// 属性字段处理
	// fieldInfos := this.table.Fields()
	// fields := this.table.FieldNames()
	// fieldCount := len(fieldInfos)
	for i := 0; i < count; i++ {
		features[i].Geo = loadFromByte(r, this.GeoType)
		// features[i].Atts = make(map[string]interface{}, fieldCount)
		// for j, name := range fields {
		// 	features[i].Atts[name] = dbfString2Value(this.table.FieldValue(i+start, j), fieldInfos[j].FieldType())
		// }
	}
}

// 一次性获得所有对象的bbox和id
func (this *ShapeFile) LoadBboxIds() (bboxes []base.Rect2D, ids []int32) {
	bboxes = make([]base.Rect2D, this.recordNum)
	ids = make([]int32, this.recordNum)

	f, _ := os.Open(this.Filename)
	// for i:=0;i<this.recordNum; i++ {
	for i, v := range this.records {
		f.Seek(int64(v.Pos)*2, 0)
		bboxes[i], ids[i] = this.loadOneBboxId(f, this.GeoType)
	}
	return
}

// 从 reader中读取一个对象的bbox和id
func (this *ShapeFile) loadOneBboxId(r io.Reader, shptype geometry.ShpType) (bbox base.Rect2D, id int32) {
	var num, len int32
	binary.Read(r, binary.BigEndian, &num)
	id = num - 1 // num是从1起的，而id应从0起
	binary.Read(r, binary.BigEndian, &len)

	switch shptype {
	case ShpPoint:
		var geopoint geometry.GeoPoint
		var geotype shpType
		binary.Read(r, binary.LittleEndian, &geotype)
		binary.Read(r, binary.LittleEndian, &geopoint.Point2D)
		bbox.Min = geopoint.Point2D
		bbox.Max = geopoint.Point2D
	case ShpPolyLine:
	case ShpPolygon:
		bbox = loadShpOnePolyBbox(r)
	}

	return
}

// 把dbf字段值，从字符串转为go语言的某个具体数据类型
// Character DbaseDataType = 'C'
// 	Logical   DbaseDataType = 'L'  // ? Y y N n T t F f (? 表示没有初始化)。
// 	Date      DbaseDataType = 'D'
// 	Numeric   DbaseDataType = 'N'
// 	Float     DbaseDataType = 'F'
const TIME_LAYOUT = "2006-01-02 15:04:05"

func dbfString2Value(str string, ftype godbf.DbaseDataType) (v interface{}) {
	switch ftype {
	case godbf.Character:
		v = str
	case godbf.Logical:
		v = (str == "Y" || str == "y" || str == "T" || str == "t")
	case godbf.Date:
		v, _ = time.Parse(TIME_LAYOUT, str)
	case godbf.Numeric:
		// v, _ := strconv.Atoi(str)
		v, _ = strconv.ParseFloat(str, 64)
	case godbf.Float:
		v, _ = strconv.ParseFloat(str, 64)
	}
	return v
}

// 从内存中读取一条记录
func loadFromByte(r io.Reader, shptype geometry.ShpType) geometry.Geometry {
	// shp文件中的记录头
	var num, len int32
	binary.Read(r, binary.BigEndian, &num)
	binary.Read(r, binary.BigEndian, &len)

	data := make([]byte, len*2) // len 的单位是双字节
	binary.Read(r, binary.LittleEndian, data)

	// fmt.Println("len:", len, "data:", data)

	geotype := geometry.ShpType2Geo(shptype)
	geo := geometry.CreateGeo(geotype)
	geo.From(data, geometry.Shape)
	// shape格式中，num是从1起的，为提高效率，这里减1，变为和[] pos一致
	geo.SetID(int64(num - 1))
	return geo

	// var geo geometry.Geometry
	// switch shptype {
	// case ShpPoint:
	// 	geo = loadShpOnePoint(r)
	// case ShpPolyLine:
	// 	geo = loadShpOnePolyline(r)
	// case ShpPolygon:
	// 	geo = loadShpOnePolygon(r)
	// }
	// // shape格式中，num是从1起的，为提高效率，这里减1，变为和[] pos一致
	// geo.SetID(int64(num - 1))
	// return geo
}

// func loadShpOnePoint(r io.Reader) geometry.Geometry {
// 	var geopoint geometry.GeoPoint
// 	var geotype shpType
// 	binary.Read(r, binary.LittleEndian, &geotype)
// 	binary.Read(r, binary.LittleEndian, &geopoint.Point2D)
// 	// fmt.Println("geopoint:", geopoint)
// 	return &geopoint
// }

// type shpPolyline struct {
// 	shpType                  // 图形类型，==3
// 	bbox       base.Rect2D    // 当前线状目标的坐标范围
// 	numParts  int32          // 当前线目标所包含的子线段的个数
// 	numPoints int32          // 当前线目标所包含的顶点个数
// 	parts     []int32        // 每个子线段的第一个坐标点在 Points 的位置
// 	points    []base.Point2D // 记录所有坐标点的数组
// }
// 未来考虑是否放到geometry的各个类型中实现
// func loadShpOnePolyline(r io.Reader) geometry.Geometry {
// 	var polyline geometry.GeoPolyline
// 	bbox, numParts, numPoints := loadShpOnePolyHeader(r)
// 	polyline.BBox = bbox

// 	parts := make([]int32, numParts, numParts+1)
// 	binary.Read(r, binary.LittleEndian, parts)
// 	parts = append(parts, numPoints) // 最后增加一个，方便后面的计算

// 	polyline.Points = make([][]base.Point2D, numParts)
// 	for i := int32(0); i < numParts; i++ {
// 		polyline.Points[i] = make([]base.Point2D, parts[i+1]-parts[i])
// 		binary.Read(r, binary.LittleEndian, polyline.Points[i])
// 	}

// 	return &polyline
// }

// func loadShpOnePolygon(r io.Reader) geometry.Geometry {
// 	var polygon geometry.GeoPolygon
// 	bbox, numParts, numPoints := loadShpOnePolyHeader(r)
// 	polygon.BBox = bbox

// 	parts := make([]int32, numParts+1)
// 	for i := int32(0); i < numParts; i++ {
// 		binary.Read(r, binary.LittleEndian, &parts[i])
// 	}
// 	parts[numParts] = numPoints

// 	polygon.Points = make([][][]base.Point2D, numParts)
// 	for i := int32(0); i < numParts; i++ {
// 		polygon.Points[i] = make([][]base.Point2D, 1)
// 		polygon.Points[i][0] = make([]base.Point2D, parts[i+1]-parts[i])
// 		binary.Read(r, binary.LittleEndian, polygon.Points[i][0])
// 	}
// 	return &polygon
// }

// 读取 bbox
func loadShpOnePolyBbox(r io.Reader) (bbox base.Rect2D) {
	var shptype int32
	binary.Read(r, binary.LittleEndian, &shptype)
	// 这里合并处理
	binary.Read(r, binary.LittleEndian, &bbox)
	return
}

// // 读取 polyline和polygon共同的部分
// func loadShpOnePolyHeader(r io.Reader) (bbox base.Rect2D, numParts, numPoints int32) {
// 	var shptype int32
// 	binary.Read(r, binary.LittleEndian, &shptype)
// 	// 这里合并处理
// 	binary.Read(r, binary.LittleEndian, &bbox)
// 	binary.Read(r, binary.LittleEndian, &numParts)
// 	binary.Read(r, binary.LittleEndian, &numPoints)
// 	return
// }
