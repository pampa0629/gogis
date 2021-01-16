package draw

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/chai2010/webp"
)

// 地图输出格式类型定义
type MapType string

const (
	// TypeBmp  MapType = "bmp" // todo
	TypePng  MapType = "png"
	TypeJpg  MapType = "jpg"
	TypeJpeg MapType = "jpeg" // == jpg
	TypeWebp MapType = "webp"
	TypeMvt  MapType = "mvt" // mapbox vector tile
	// TypePdf  MapType = "pdf" // todo adobe pdf
)

// 返回是否为图片格式
func (this *MapType) IsImgType() bool {
	switch *this {
	// case TypeBmp:
	case TypePng, TypeJpg, TypeJpeg, TypeWebp:
		return true
	}
	return false
}

func (this *MapType) OutputImg(w io.Writer, m image.Image) {
	switch *this {
	case "png":
		png.Encode(w, m)
	case "jpg", "jpeg":
		jpeg.Encode(w, m, nil)
	case "webp":
		webp.Encode(w, m, nil)
	default:
		fmt.Println("不支持的图片格式：", this)
	}
}

func (this *MapType) OutputImg2Bytes(m image.Image) []byte {
	data := make([]byte, 0)
	buf := bytes.NewBuffer(data)
	this.OutputImg(buf, m)
	return buf.Bytes()
}
