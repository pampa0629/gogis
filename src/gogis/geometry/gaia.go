package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
)

type GAIAInfo struct {
	// static byte byteOrdering = 1; //字节序：小端存储
	// int32 srid; //坐标系 ID
	// Rect mbr; //对象的坐标范围
	// static byte gaiaMBR=0x7c; //MBR 结束标识
	order byte  // 字节顺序
	srid  int32 // 坐标系id
	bbox  base.Rect2D
	mark  byte
}

func (this *GAIAInfo) From(buf *bytes.Buffer) binary.ByteOrder {
	binary.Read(buf, binary.LittleEndian, &this.order)
	byteOrder := GAIAByte2Order(this.order)
	binary.Read(buf, byteOrder, &this.srid)
	binary.Read(buf, byteOrder, &this.bbox)
	binary.Read(buf, byteOrder, &this.mark)
	if this.mark != byte(0X7C) {
		panic("gaia info mark error:" + string(this.mark))
	}
	return byteOrder
}

// 字节顺序，一个字节存储
type GAIAByteOrder byte

const (
	GaiaBig    GAIAByteOrder = 0 // 大尾端
	GaiaLittle GAIAByteOrder = 1 // 小尾端，默认都用小端
)

// 读取一个字节，确定 wkb后续内容的字节顺序
func GAIAByte2Order(gaiaByte byte) (byteOrder binary.ByteOrder) {
	if gaiaByte == byte(GaiaBig) {
		byteOrder = binary.BigEndian
	} else if gaiaByte == byte(GaiaLittle) {
		byteOrder = binary.LittleEndian
	} else {
		panic("gaia byte order error:" + string(gaiaByte))
	}
	return
}
