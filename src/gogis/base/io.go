// 输入输出，类型转化等功能函数
package base

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"unsafe"
)

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
// func xxByteStruct(value interface{}, size int) (data []byte) {
// 	sv := reflect.ValueOf(value)
// 	fmt.Println("struct kind is:", sv.Kind())
// 	if sv.Kind() != reflect.Struct {
// 		panic(fmt.Sprintf("ByteStruct called with non-struct value of type %T", value))
// 	}
// 	h := (*reflect.SliceHeader)((unsafe.Pointer(&data)))
// 	// h.Cap = sv.Cap() * int(sv.Type().Elem().Size())
// 	h.Cap = size //  (int)(unsafe.Sizeof(value))
// 	h.Len = size //  (int)(unsafe.Sizeof(value))
// 	h.Data = (uintptr)(unsafe.Pointer(&value))
// 	fmt.Println("struct len is:", h.Len)
// 	return
// }

// 大小端互换
func ExEndian(value int32) int32 {
	buf := (*[4]byte)(unsafe.Pointer(&value))
	buf[0], buf[3] = buf[3], buf[0]
	buf[1], buf[2] = buf[2], buf[1]
	return value
}

// 大小端互换
func ExEndianShort(value int16) int16 {
	buf := (*[2]byte)(unsafe.Pointer(&value))
	buf[0], buf[1] = buf[1], buf[0]
	return value
}

// 大小端互换
func ExEndianDouble(value float64) float64 {
	buf := (*[8]byte)(unsafe.Pointer(&value))
	buf[0], buf[7] = buf[7], buf[0]
	buf[1], buf[6] = buf[6], buf[1]
	buf[2], buf[5] = buf[5], buf[2]
	buf[3], buf[4] = buf[4], buf[3]
	return value
}

// 从切片中读取数据
func ToInt32(input []byte, pos uint64, change bool) (data int32) {
	bytesBuffer := bytes.NewBuffer(input[pos : pos+4])
	if change {
		binary.Read(bytesBuffer, binary.BigEndian, &data)
	} else {
		binary.Read(bytesBuffer, binary.LittleEndian, &data)
	}
	return
}

// 从切片中读取数据
func ToFloat64(input []byte, pos uint64) (data float64) {
	bytesBuffer := bytes.NewBuffer(input[pos : pos+8])
	binary.Read(bytesBuffer, binary.LittleEndian, &data)
	return
}

// 从文件中读取数据
func ReadInt32(f *os.File, change bool) (data int32) {
	p := (*[4]byte)(unsafe.Pointer(&data))
	// n, err := f.Read((*p)[:]) 为啥两种写法都OK呢？
	_, _ = f.Read(p[:])
	if change {
		data = ExEndian(data)
	}
	return
}

// 从文件中读取数据
func ReadFloat64(f *os.File) (data float64) {
	p := (*[8]byte)(unsafe.Pointer(&data))
	// n, err := f.Read((*p)[:]) 为啥两种写法都OK呢？
	_, _ = f.Read(p[:])
	return
}

// ===========================================================//
// 各类 基本数据类型与bytes之间的相互转换
// ===========================================================//

func Float32ToBytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

func BytesToFloat32(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	return math.Float32frombits(bits)
}

func Float64ToBytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

func BytesToFloat64(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}

func Int32ToBytes(i int32) []byte {
	var buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(i))
	return buf
}

func BytesToInt32(buf []byte) int32 {
	return int32(binary.LittleEndian.Uint32(buf))
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.LittleEndian.Uint64(buf))
}

// ==================================================== //

// int32 转化为 bytes
// func Int2Bytes(value int32) []byte {
// 	var buf bytes.Buffer

// 	// 数字转 []byte
// 	binary.Write(&buf, binary.LittleEndian, value)
// 	return buf.Bytes()
// }

// // bytes 转为 int32
// func Bytes2Int(data []byte) (value int32) {
// 	buf := bytes.NewBuffer(data)
// 	binary.Read(buf, binary.LittleEndian, &value)
// 	return
// }

// // double 转为 bytes
// func Double2Bytes(value float64) []byte {
// 	var buf bytes.Buffer

// 	// 数字转 []byte
// 	binary.Write(&buf, binary.LittleEndian, value)
// 	return buf.Bytes()
// }

// 布尔值转化为int32，方便后面对齐存储
func BoolToInt32(value bool) (result int32) {
	if value {
		result = 1
	}
	return result
}

func CopyFile(dstFileName string, srcFileName string) (written int64, err error) {
	srcFile, err := os.Open(srcFileName)
	PrintError("OpenFile", err)
	defer srcFile.Close()

	//打开dstFileName
	dstFile, err := os.OpenFile(dstFileName, os.O_WRONLY|os.O_CREATE, 0755)
	PrintError("CreateFile", err)
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}
