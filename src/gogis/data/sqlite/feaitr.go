// 基于sqlite实现的空间文件引擎
package sqlite

import (
	"database/sql"
	"gogis/base"
	"gogis/data"
	"gogis/geometry"
	"strings"

	// "database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// 迭代器
type SqliteFeaItr struct {
	feaset  *SqliteFeaset
	bbox    base.Rect2D
	count   int64
	codes   []int32
	codess  [][]int32 // 每个批次所对应的index codes
	geotype int       // gaia 的geo类型
	fields  []string  // 要哪些字段
	where   string    // sql where
	squery  data.SpatailQuery

	countPerGo int // 每一个批次的对象数量
}

func (this *SqliteFeaItr) Count() int64 {
	return this.count
}

func (this *SqliteFeaItr) Close() {
	this.codes = this.codes[:0]
	this.codess = this.codess[:0]
	this.fields = this.fields[:0]
}

// 为了批量读取做准备，返回批量的次数
func (this *SqliteFeaItr) BeforeNext(objCount int) (goCount int) {
	if len(this.codes) > 0 {
		goCount = int(this.count)/objCount + 1
		// 这里假设每个code中所包含的对象，是大体平均分布的
		this.codess = base.SplitSlice32(this.codes, goCount)
		// fmt.Println("codes:", this.codes)
		// fmt.Println("codess:", this.codess)
		this.countPerGo = objCount
	} else {
		// 没有空间查询的编码，怎么划分，再研究 todo
		goCount = 1
	}

	return
}

func buildWhere(codes []int32, where string) (out string) {
	in := ""
	// 先处理可能的空间索引编码
	if codes != nil {
		// 如果编码存在，且长度为0，则说明啥也不能要
		if len(codes) == 0 {
			codes = []int32{-1}
		}
		in = "g_index_code" + buildSqlIn(codes)
		out += in
	}
	// 有用户输入的where内容，再加上去
	if len(where) > 0 {
		// 两个都有时，还需要加  and ( )
		if len(in) > 0 {
			where = " and (" + where + ")"
		}
		out += where
	}
	if len(out) > 0 {
		out = " where " + out
	}
	return
}

// 构造 选择 语句
func (this *SqliteFeaItr) buildSelect() string {
	sel := " select "
	// 选择特定字段
	if this.fields != nil {
		sel += this.feaset.id + "," + this.feaset.geom
		for _, v := range this.fields {
			sel += ", " + v
		}
	} else {
		// 选择所有字段
		sel += " * "
	}
	return sel
}

// 读取某个批次的所有数据
func (this *SqliteFeaItr) BatchNext(batchNo int) (feas []data.Feature, result bool) {
	db, err := sql.Open("sqlite3", this.feaset.store.filename)
	base.PrintError("db open:"+this.feaset.store.filename, err)
	if err == nil && db != nil {
		sel := this.buildSelect()
		from := " from " + this.feaset.Name
		codes := []int32{}
		if this.codess != nil && batchNo < len(this.codess) {
			codes = this.codess[batchNo]
		}
		where := buildWhere(codes, this.where)
		sql := sel + from + where
		rows, err := db.Query(sql)
		base.PrintError("db query, sql:"+sql, err)
		if err == nil {
			result = true
			geotype := SPL2GeoType(this.geotype)
			feas = make([]data.Feature, 0, this.countPerGo)
			for rows.Next() {
				fea := fetchFea(rows, geotype, this.feaset.geom, this.feaset.id)
				if fea.Geo != nil && this.squery.Match(fea.Geo) {
					feas = append(feas, fea)
				}
			}
			defer rows.Close()
		}
		defer db.Close()
	}
	return
}

// 获取feature
func fetchFea(rows *sql.Rows, geoType geometry.GeoType, geom, id string) (fea data.Feature) {
	cols, _ := rows.Columns()
	colsCount := len(cols)
	// fmt.Println("cols count:", colsCount)
	atts := make([]interface{}, colsCount)
	for i, _ := range atts {
		atts[i] = new(interface{})
	}
	err := rows.Scan(atts...)
	base.PrintError("fetch feature", err)
	if err == nil {
		fea.Geo = geometry.CreateGeo(geoType)
		if colsCount > 2 {
			fea.Atts = make(map[string]interface{})
		}
		for i, v := range cols {
			v = strings.ToLower(v)
			att := *atts[i].(*interface{})
			if att != nil {
				switch v {
				case id:
					fea.Geo.SetID(att.(int64))
				case geom:
					fea.Geo.From(att.([]byte), geometry.GAIA)
				case "g_index_code":
					// do nothing
				default:
					fea.Atts[v] = att
				}
			}
		}
	}
	return
}
