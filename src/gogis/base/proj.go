// 投影类
package base

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	// "github.com/go-spatial/proj"
	"github.com/go-spatial/proj"
)

var goProj Proj

func init() {
	// 加载 proj.db
	goProj.Load()
}

// 空间参考系的全部信息，epsg、proj4和wkt，三者有一个即可确定
type ProjInfo struct {
	Epsg  int
	Name  string
	Proj4 string // proj4 格式描述
	Wkt   string // wkt格式描述，同时也是arcgis的描述
}

// 是否为投影系
func (this *ProjInfo) IsProjection() bool {
	// 找到 longlat 即为 经纬坐标系统，也就是非投影系统
	pos := strings.Index(this.Proj4, "longlat")
	return pos < 0
}

type Proj struct {
	projInfos []*ProjInfo
}

func (this *Proj) Load() {
	fmt.Println("app path:", os.Args[0])
	dbPath := GetAbsolutePath(os.Args[0], "./proj.db")
	if !IsExist(dbPath) {
		dbPath = GetAbsolutePath(os.Args[0], "../../../proj.db")
	}
	db, err := sql.Open("sqlite3", dbPath)
	PrintError("open proj db", err)
	sql := "select srid,ref_sys_name,proj4text,srtext from spatial_ref_sys"
	rows, err := db.Query(sql)
	PrintError("query proj db, sql is:"+sql, err)
	for i := 0; rows.Next(); i++ {
		var oneProj ProjInfo
		err := rows.Scan(&oneProj.Epsg, &oneProj.Name, &oneProj.Proj4, &oneProj.Wkt)
		if err == nil {
			oneProj.Name = strings.TrimSpace(oneProj.Name)
			oneProj.Proj4 = strings.TrimSpace(oneProj.Proj4)
			oneProj.Wkt = strings.TrimSpace(oneProj.Wkt)
			this.projInfos = append(this.projInfos, &oneProj)
		}
	}
	rows.Close() // 记得关闭
}

// 根据 wkt string 得到 proj info
func PrjFromEpsg(epsg int) *ProjInfo {
	for _, v := range goProj.projInfos {
		if v.Epsg == epsg {
			return v
		}
	}
	return nil
}

// 根据 wkt string 得到 proj info
func PrjFromWkt(wktStr string) *ProjInfo {
	if len(wktStr) > 0 {
		var wkt WktProj
		wkt.Parse(wktStr)
		// 若能通过匹配，找到对应的epsg；则通过epsg，得到proj4
		projInfo := goProj.matchEpsg(&wkt)
		if projInfo == nil {
			// 若匹配不上，则epsg为自定义，并根据内容输出 proj4
			projInfo = new(ProjInfo)
			projInfo.Proj4 = wkt.toProj4()
		}
		return projInfo
	}
	return nil
}

func (this *Proj) matchEpsg(wkt *WktProj) *ProjInfo {
	// "GCS_WGS_1984" "WGS 84"
	// "GCS_China_2000" "China Geodetic Coordinate System 2000	"
	// 匹配的方法:去掉GCS_等前缀,把_19换为空格;再与name进行对比
	name := strings.Replace(wkt.values[0].(string), "GCS_", "", 1)
	name = strings.Replace(name, "_19", " ", -1)
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Replace(name, "\"", "", -1)
	for _, v := range this.projInfos {
		if v.Name == name {
			return v
		} else {
			// 看是否能包括所有关键字
			findit := true
			items := strings.Split(name, " ")
			for _, item := range items {
				if strings.Index(v.Name, item) < 0 {
					findit = false // 有一个没找到，就不酸
					break
				}
			}
			if findit {
				return v
			}
		}
	}
	return nil
}

type WktProj struct {
	name   string
	values []interface{}
}

// todo
// 根据自身内容，输出proj4格式
func (this *WktProj) toProj4() string {
	return ""
}

// GEOGCS[	"GCS_WGS_1984",
// 	DATUM[	"D_WGS_1984",
// 		SPHEROID["WGS_1984",6378137,298.257223563]],
// 	PRIMEM["Greenwich",0],
// 	UNIT["Degree",0.017453292519943295]
func (this *WktProj) Parse(wkt string) {
	// 找到第一个[，之前的部分为 name;
	// 第一个 [ 和最后一个 ] 中间的部分为 数组，通过查找 , 和 [] 的关系,分解为多个string, 生成 values
	// 判断 string,看 value为 数字(整数或浮点), 字符串 或者 下一级的 WktProj
	pos := strings.Index(wkt, "[")
	this.name = wkt[0:pos]
	value := strings.Trim(wkt[pos:], "[] ")
	// fmt.Println("value:", value)
	values := this.splitValues(value)
	for _, v := range values {
		switch v[0] {
		case '"':
			this.values = append(this.values, v)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if strings.Index(v, ".") >= 0 {
				value, _ := strconv.ParseFloat(v, 64)
				this.values = append(this.values, value)
			} else {
				value, _ := strconv.ParseInt(v, 10, 64)
				this.values = append(this.values, value)
			}
		default:
			newWkt := new(WktProj)
			newWkt.Parse(v)
			this.values = append(this.values, newWkt)
		}
	}
}

func (this *WktProj) splitValues(wkt string) (values []string) {
	last, indent := 0, 0
	for i := 0; i < len(wkt); i++ {
		switch wkt[i] {
		case ',':
			if indent == 0 {
				values = append(values, strings.Trim(wkt[last:i], " "))
				last = i + 1 // 跳过 ","
			}
		case '[':
			indent++
		case ']':
			indent--
		}
	}
	values = append(values, wkt[last:len(wkt)])
	return
}

type PrjConvert struct {
	from *ProjInfo
	to   *ProjInfo
}

// from 和 to 都不得为nil，否则返回 nil
func NewPrjConvert(from, to *ProjInfo) (prjc *PrjConvert) {
	if from != nil && to != nil {
		prjc = new(PrjConvert)
		prjc.from = from
		prjc.to = to
		// 对于投影系统，则先注册进去
		if prjc.from.IsProjection() {
			proj.Register(proj.EPSGCode(prjc.from.Epsg), prjc.from.Proj4)
		}
		if prjc.to.IsProjection() {
			proj.Register(proj.EPSGCode(prjc.to.Epsg), prjc.to.Proj4)
		}
	}
	return
}

func (this *PrjConvert) DoPnt(pnt Point2D) (dest Point2D) {
	pnts := make([]Point2D, 1)
	pnts[0] = pnt
	return this.DoPnts(pnts)[0]
}

func (this *PrjConvert) DoPnts(pnts []Point2D) (dest []Point2D) {
	// 具体分为四种情况，分别为：
	// 1）经纬 转 经纬，无需转化
	if !this.from.IsProjection() && !this.to.IsProjection() {
		dest = pnts
		return
	}

	// 既然肯定要转，先准备数据
	src := make([]float64, len(pnts)*2)
	for i, v := range pnts {
		src[i*2] = v.X
		src[i*2+1] = v.Y
	}

	// 2）投影 转 经纬，只用from转一次
	if this.from.IsProjection() {
		src, _ = proj.Inverse(proj.EPSGCode(this.from.Epsg), src)
	}
	// 3）经纬 转 投影，只用to转一次
	if this.to.IsProjection() {
		src, _ = proj.Convert(proj.EPSGCode(this.to.Epsg), src)
	}
	// 4）投影 转 投影，from和to都要用，转两次
	dest = make([]Point2D, len(pnts))
	for i, _ := range dest {
		dest[i].X = src[i*2]
		dest[i].Y = src[i*2+1]
	}
	return
}
