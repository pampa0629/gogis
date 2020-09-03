package data

// 字符解析，性能都不理想，暂时搁置

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/axgle/mahonia"
)

// dbf 文件头
type DbfHeader struct {
	Version    byte
	Data       [3]byte // Y M D
	RecordNum  int32
	HeaderSize int16
	RecordSize int16
	Unused     [20]byte
}

func (this *DbfHeader) read(r io.Reader) {
	binary.Read(r, binary.LittleEndian, this)
	fmt.Println(this)
}

// 字段描述信息
type DbfFieldDesc struct {
	Name          [11]byte
	Type          byte
	Unused1       [4]byte
	Length        byte
	DecimalPlaces byte
	Unused2       [14]byte
}

func readDbfFieldDesc(r io.Reader, encoder mahonia.Encoder) (finfo FieldInfo) {
	var fieldDesc DbfFieldDesc
	binary.Read(r, binary.LittleEndian, &fieldDesc)
	fmt.Println(fieldDesc)

	finfo.Name = strings.Trim(encoder.ConvertString(string(fieldDesc.Name[:])), string([]byte{0}))
	finfo.Length = int(fieldDesc.Length)
	finfo.Type = dbfTypeConvertor2(fieldDesc.Type)
	return
}

// (B, C, D, N, L, M, @, I, +, F, 0 or G).
func dbfTypeConvertor2(dtype byte) FieldType {
	ftype := TypeUnknown
	switch dtype {
	case 'C':
		ftype = TypeString
	case 'L':
		ftype = TypeBool
	case 'D':
		ftype = TypeTime
	case 'N': // 字符串数字，也可以是浮点数
		ftype = TypeFloat
	case 'F':
		ftype = TypeFloat
	}
	return ftype
}

type DbfFile struct {
	// filename string //  文件名
	DbfHeader
	FieldNum   int             // 字段个数
	fieldInfos []FieldInfo     // 字段信息
	records    [][]interface{} // 实际内容
	encoder    mahonia.Encoder
	data       []byte
}

func (this *DbfFile) Open(filename string, encoding string) {
	// 字符编码转化器
	this.encoder = mahonia.NewEncoder(encoding)

	// 一次性加载到内存
	this.data, _ = ioutil.ReadFile(filename)
	rh := bytes.NewReader(this.data)

	// 读取文件头
	this.DbfHeader.read(rh)

	// 读取字段信息
	this.FieldNum = (int)((this.HeaderSize - 1 - 32) / 32)
	this.fieldInfos = make([]FieldInfo, this.FieldNum)
	for i := 0; i < this.FieldNum; i++ {
		this.fieldInfos[i] = readDbfFieldDesc(rh, this.encoder)
		fmt.Println(this.fieldInfos[i])
	}

	this.records = make([][]interface{}, this.RecordNum)
	forcount := (int)(this.RecordNum/ONE_LOAD_COUNT) + 1
	fmt.Println("DbfFile.Open(), for count: ", forcount)
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < forcount; i++ {
		count := ONE_LOAD_COUNT
		if i == forcount-1 { // 最后一次循环，剩余的对象个数
			count = (int(this.RecordNum) - ONE_LOAD_COUNT*(forcount-1))
		}
		wg.Add(1)
		go this.BatchLoad(i*ONE_LOAD_COUNT, count, wg)
	}
	wg.Wait()
}

func (this *DbfFile) BatchLoad(start int, count int, wg *sync.WaitGroup) {
	defer wg.Done()

	offset := int(this.HeaderSize) + start*int(this.RecordSize)
	rr := bytes.NewReader(this.data[offset:])
	// 读取记录
	for i := 0; i < count; i++ {
		rr.Seek(1, 0) // 记录头一个字节跳过
		this.records[i+start] = this.readOneRecord(rr)
		fmt.Println(this.records[i])
	}
}

func (this *DbfFile) readOneRecord(r io.Reader) (record []interface{}) {
	record = make([]interface{}, this.FieldNum)
	for i, finfo := range this.fieldInfos {
		data := make([]byte, finfo.Length)
		r.Read(data)
		record[i] = this.dbfString2Value2(data, finfo.Type)
	}
	return
}

func (this *DbfFile) dbfString2Value2(data []byte, ftype FieldType) (v interface{}) {
	switch ftype {
	case TypeString:
		str := strings.Trim(this.encoder.ConvertString(string(data[:])), string([]byte{0}))
		v = strings.TrimSpace(str)
	case TypeBool:
		v = (data[0] == 'Y' || data[0] == 'y' || data[0] == 'T' || data[0] == 't')
	case TypeTime:
		str := strings.Trim(string(data[:]), string([]byte{0}))
		v, _ = time.Parse(TIME_LAYOUT, str)
	case TypeFloat:
		str := strings.Trim(string(data[:]), string([]byte{0}))
		v, _ = strconv.ParseFloat(str, 64)
	}
	return v
}
