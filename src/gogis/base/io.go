package base

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
