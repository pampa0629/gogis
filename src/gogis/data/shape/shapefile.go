package shape

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gogis/base"
	"gogis/data"
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
	prj       string // 坐标系的字符串描述(wkt格式)
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
	if this.table != nil {
		this.table.Close()
	}
}

func (this *ShapeFile) Open(filename string, loadFields bool) bool {
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

	// prj 文件
	prjName := strings.TrimSuffix(filename, ".shp") + ".prj"
	prjData, _ := ioutil.ReadFile(prjName)
	this.prj = string(prjData)

	if loadFields {
		// dbf 文件
		dbfName := strings.TrimSuffix(filename, ".shp") + ".dbf"
		this.table, _ = godbf.NewFromFile(dbfName, "UTF8")
	}

	return true
}

func (this *ShapeFile) GetFieldInfos() (finfos []data.FieldInfo) {
	if this.table != nil {
		fds := this.table.Fields()
		finfos = make([]data.FieldInfo, len(fds))
		for i, ds := range fds {
			finfos[i].Name = ds.Name()
			finfos[i].Length = int(ds.Length())
			finfos[i].Type = dbfTypeConvertor(ds.FieldType())
		}
	}
	return
}

func dbfTypeConvertor(dtype godbf.DbaseDataType) data.FieldType {
	ftype := data.TypeUnknown
	switch dtype {
	case godbf.Character:
		ftype = data.TypeString
	case godbf.Logical:
		ftype = data.TypeBool
	case godbf.Date:
		ftype = data.TypeTime
	case godbf.Numeric: // 字符串数字，也可以是浮点数
		ftype = data.TypeFloat
	case godbf.Float:
		ftype = data.TypeFloat
	}
	return ftype
}

// 读取一条记录
// func (this *ShapeFile) LoadOne(f *os.File, id int) (feature data.Feature) {

// 	// 先确定要读取文件的位置和长度
// 	pos := uint64(this.records[id].Pos) * 2
// 	// length := this.records[id].Length*2 + 8

// 	// f, err := os.Open(this.Filename)
// 	// if err != nil {
// 	// 	fmt.Println("open shape file error:", err.Error())
// 	// }
// 	f.Seek(int64(pos), 0)

// 	feature.Geo = loadFromByte(f, this.GeoType)
// 	return

// 	// 属性字段处理
// 	// fieldInfos := this.table.Fields()
// 	// fields := this.table.FieldNames()
// 	// fieldCount := len(fieldInfos)
// 	// for i := 0; i < count; i++ {

// 	// features[i].Atts = make(map[string]interface{}, fieldCount)
// 	// for j, name := range fields {
// 	// 	features[i].Atts[name] = dbfString2Value(this.table.FieldValue(i+start, j), fieldInfos[j].FieldType())
// 	// }
// 	// }
// }

// 加载start位置开始，批量读取geometry
// 要求读取的count都是连续存储的
func (this *ShapeFile) BatchLoad(f *os.File, start int, count int, geos []geometry.Geometry, wg *sync.WaitGroup) {
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
	}

	f.Seek(int64(pos), 0)
	data := make([]byte, length)
	f.Read(data)
	// fmt.Println("read shp  count:", n, " error:", err)
	r := bytes.NewBuffer(data)

	for i := 0; i < count; i++ {
		geos[i] = loadFromByte(r, this.GeoType)
	}
}

// func (this *ShapeFile) BatchLoad(f *os.File, start int, count int, features []data.Feature, fields []string, wg *sync.WaitGroup) {
// 	// fmt.Println("ShapeFile.BatchLoad()", start, count, this.geoType, this.Filename)
// 	if wg != nil {
// 		defer wg.Done()
// 	}

// 	// 先确定要读取文件的位置和长度
// 	pos := uint64(this.records[start].Pos) * 2
// 	length := int32(0)
// 	end := start + count
// 	for i := start; i < end; i++ {
// 		length += this.records[i].Length*2 + 8 // 8个字节是 shp记录的头
// 	}

// 	if f == nil {
// 		f, _ = os.Open(this.Filename)
// 		defer f.Close()
// 		// if err != nil {
// 		// 	fmt.Println("open shape file error:", err.Error())
// 		// }
// 	}

// 	f.Seek(int64(pos), 0)
// 	data := make([]byte, length)
// 	f.Read(data)
// 	// fmt.Println("read shp  count:", n, " error:", err)
// 	r := bytes.NewBuffer(data)

// 	// 属性字段处理
// 	fieldInfos := this.getDbfFields()
// 	fieldCount := len(fields)
// 	findexs := make([]int, fieldCount)
// 	ftypes := make([]godbf.DbaseDataType, fieldCount)
// 	for i, v := range fields {
// 		for j, finfo := range fieldInfos {
// 			if v == finfo.Name() {
// 				findexs[i] = j
// 				ftypes[i] = finfo.FieldType()
// 				break
// 			}
// 		}
// 	}

// 	for i := 0; i < count; i++ {
// 		features[i].Geo = loadFromByte(r, this.GeoType)
// 		if this.table != nil {
// 			features[i].Atts = make(map[string]interface{}, fieldCount)
// 			for j := 0; j < fieldCount; j++ {
// 				features[i].Atts[fields[j]] = dbfString2Value(this.table.FieldValue(i+start, j), ftypes[j])
// 			}
// 		}
// 	}
// }

func (this *ShapeFile) getDbfFields() (infos []godbf.FieldDescriptor) {
	if this.table != nil {
		infos = this.table.Fields()
	}
	return infos
}

// 一次性获得所有对象的bbox和id
func (this *ShapeFile) LoadBboxIds() (bboxes []base.Rect2D, ids []int32) {
	bboxes = make([]base.Rect2D, this.recordNum)
	ids = make([]int32, this.recordNum)

	f, _ := os.Open(this.Filename)
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
	case ShpPolyLine, ShpPolygon:
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

func dbfString2Value(str string, ftype godbf.DbaseDataType) (v interface{}) {
	switch ftype {
	case godbf.Character:
		v = str
	case godbf.Logical:
		v = (str == "Y" || str == "y" || str == "T" || str == "t")
	case godbf.Date:
		v, _ = time.Parse(data.TIME_LAYOUT, str)
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

	geotype := geometry.ShpType2Geo(shptype)
	geo := geometry.CreateGeo(geotype)
	geo.From(data, geometry.Shape)
	// shape格式中，num是从1起的，为提高效率，这里减1，变为和[] pos一致
	geo.SetID(int64(num - 1))
	return geo
}

// 读取 bbox
func loadShpOnePolyBbox(r io.Reader) (bbox base.Rect2D) {
	var shptype int32
	binary.Read(r, binary.LittleEndian, &shptype)
	// 这里合并处理
	binary.Read(r, binary.LittleEndian, &bbox)
	return
}
