package geometry

import (
	"bytes"
	"encoding/binary"
	"gogis/base"
	"gogis/draw"
	"io"
)

type GeoPolyline struct {
	Points [][]base.Point2D
	BBox   base.Rect2D
	GeoID
}

func (this *GeoPolyline) Type() GeoType {
	// 	if len(this.Points) == 1 {
	// 		return TGeoLineString
	// 	} else if len(this.Points) > 1 {
	// 		return TGeoMultiLineString
	// 	}
	return TGeoPolyline
}

func (this *GeoPolyline) Clone() Geometry {
	res := new(GeoPolyline)
	res.id = this.id
	res.BBox = this.BBox
	res.Points = make([][]base.Point2D, len(this.Points))
	for i, v := range this.Points {
		res.Points[i] = make([]base.Point2D, len(v))
		for ii, vv := range v {
			res.Points[i][ii] = vv
		}
	}
	return res
}

func (this *GeoPolyline) GetBounds() base.Rect2D {
	return this.BBox
}

func (this *GeoPolyline) ComputeBounds() base.Rect2D {
	this.BBox.Init()
	for _, points := range this.Points {
		this.BBox.Union(base.ComputeBounds(points))
	}
	return this.BBox
}

func (this *GeoPolyline) Draw(canvas *draw.Canvas) {
	var line = new(draw.Polyline)
	line.Points = make([][]draw.Point, len(this.Points))
	for i, v := range this.Points {
		line.Points[i] = canvas.Params.Forwards(v)
	}
	canvas.DrawPolyPolyline(line)
}

func (this *GeoPolyline) ConvertPrj(prjc *base.PrjConvert) {
	if prjc != nil {
		for i, v := range this.Points {
			this.Points[i] = prjc.Do(v)
		}
	}
}

// 抽稀一条折线
func thinOneLine(points []base.Point2D, dis2 float64) (newPnts []base.Point2D) {
	newPnts = make([]base.Point2D, 1, len(points))
	// dis2 := math.Pow(dis, 2)
	newPnts[0] = points[0]
	pos := 0

	count := len(points)
	for i := 1; i < count; i++ {
		// 距离够远，或者拐角够大的点，都应该保留下来
		if base.DistanceSquare(points[pos].X, points[pos].Y, points[i].X, points[i].Y) > dis2 ||
			(i < count-1 && base.Angle(points[pos], points[i], points[i+1]) < 120) {
			newPnts = append(newPnts, points[i])
			pos = i
		}
	}
	// 点数不够，就重复一下
	for len(newPnts) < 2 {
		newPnts = append(newPnts, newPnts[0])
	}
	// 点数小于2，则返回nil
	// if len(newPnts) < 2 {
	// 	return nil
	// }
	return newPnts
}

func (this *GeoPolyline) Thin(dis2 float64) Geometry {
	var newgeo GeoPolyline
	newgeo.BBox = this.BBox
	newgeo.Points = make([][]base.Point2D, 0, len(this.Points))
	for _, v := range this.Points {
		pnts := thinOneLine(v, dis2)
		if pnts != nil {
			newgeo.Points = append(newgeo.Points, pnts)
		}
	}
	if len(newgeo.Points) == 0 {
		return nil
	}
	return &newgeo
}

// wkb:
// WKBLineString
// byte  byteOrder;                               //字节序
// static  uint32  wkbType = 2;                    //几何体类型
// uint32  numPoints;                                  //点的个数
// Point  points[numPoints]}                 //点的坐标数组

// WKBMultiLineString
// byte  byteOrder;                                                    //字节序
// staticuint32 wkbType = 5;                                       //几何体类型
// uint32  numLineStrings;                                        //线串的个数
// WKBLineString  lineStrings[numLineStrings]}         //线串数组

func (this *GeoPolyline) From(data []byte, mode GeoMode) bool {
	this.Points = this.Points[0:0] // 清空
	buf := bytes.NewBuffer(data)

	switch mode {
	case WKB:
		return this.fromWkb(buf)
	// case WKT: // todo
	case Shape:
		return this.fromShp(buf)
	}
	return false
}

func (this *GeoPolyline) fromWkb(r io.Reader) bool {
	var order byte
	binary.Read(r, binary.LittleEndian, &order)
	byteOrder := WkbByte2Order(order)
	var wkbType WkbGeoType
	binary.Read(r, byteOrder, &wkbType)
	if wkbType == WkbLineString { // 简单线
		this.Points = append(this.Points, Bytes2WkbLinearRing(r, byteOrder))
	} else if wkbType == WkbMultiLineString { // 复杂线
		var numLineStrings uint32
		binary.Read(r, byteOrder, &numLineStrings)
		this.Points = make([][]base.Point2D, numLineStrings)
		for i := uint32(0); i < numLineStrings; i++ {
			var order byte
			binary.Read(r, binary.LittleEndian, &order)
			byteOrder := WkbByte2Order(order)
			var wkbType WkbGeoType
			binary.Read(r, byteOrder, &wkbType)
			this.Points[i] = Bytes2WkbLinearRing(r, byteOrder)
		}
	}
	this.ComputeBounds()
	return true
}

// type shpPolyline struct {
// 	shpType                  // 图形类型，==3
// 	bbox       base.Rect2D    // 当前线状目标的坐标范围
// 	numParts  int32          // 当前线目标所包含的子线段的个数
// 	numPoints int32          // 当前线目标所包含的顶点个数
// 	parts     []int32        // 每个子线段的第一个坐标点在 Points 的位置
// 	points    []base.Point2D // 记录所有坐标点的数组
// }
func (this *GeoPolyline) fromShp(r io.Reader) bool {
	// var polyline geometry.GeoPolyline
	bbox, numParts, numPoints := loadShpOnePolyHeader(r)
	this.BBox = bbox

	parts := make([]int32, numParts, numParts+1)
	binary.Read(r, binary.LittleEndian, parts)
	parts = append(parts, numPoints) // 最后增加一个，方便后面的计算

	this.Points = make([][]base.Point2D, numParts)
	for i := int32(0); i < numParts; i++ {
		this.Points[i] = make([]base.Point2D, parts[i+1]-parts[i])
		binary.Read(r, binary.LittleEndian, this.Points[i])
	}
	return true
}

func (this *GeoPolyline) To(mode GeoMode) []byte {
	switch mode {
	case WKB:
		return this.toWkb()
	case WKT:
		// todo
	}
	return nil
}

func (this *GeoPolyline) toWkb() []byte {
	var buf bytes.Buffer
	if len(this.Points) == 1 { // 简单线
		WkbLineString2Bytes(this.Points[0], &buf)
	} else { // 复杂线
		binary.Write(&buf, binary.LittleEndian, byte(WkbLittle))
		binary.Write(&buf, binary.LittleEndian, uint32(WkbMultiLineString))
		binary.Write(&buf, binary.LittleEndian, uint32(len(this.Points)))
		for _, v := range this.Points {
			WkbLineString2Bytes(v, &buf)
		}
	}
	return buf.Bytes()
}
