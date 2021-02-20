// 基于sqlite实现的空间文件引擎
package sqlite

import (
	"database/sql"
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/geometry"
	"gogis/index"
	"os"
	"strconv"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// 矢量数据集合
type SqliteFeaset struct {
	data.FeasetInfo
	id         string               // id用的字段，默认为pk，udbx中的是SmID
	geom       string               // geometry用的字段，默认为"geom"
	count      int64                // 对象个数
	indexLevel int32                // 索引层级
	idx        index.SpatialIndexDB // 暂时只支持 zorder索引 ZOrderIndex XzorderIndex
	store      *SqliteStore
	lock       sync.Mutex
}

func (this *SqliteFeaset) Open() (bool, error) {
	this.id = "pk" // 默认为pk
	// 加载统计信息，没有的话就先做统计
	this.loadGeoColsSta()
	// 读取字段信息
	this.loadGeoColsFields()
	// 再判断 空间索引是否存在，不存在 则要创建之
	this.initSpatailIndex()
	return true, nil
}

// 加载所有字段信息
func (this *SqliteFeaset) loadGeoColsFields() {
	rows, err := this.store.db.Query("select * from " + this.Name + " limit 0")
	base.PrintError("loadGeoColsFields", err)
	columns, err := rows.ColumnTypes()
	base.PrintError("ColumnTypes", err)
	this.FieldInfos = this.FieldInfos[:0]
	for i, v := range columns {
		lower := strings.ToLower(v.Name())
		geom := strings.ToLower(this.geom)
		keys := []string{"pk", "smid", geom, "g_index_code"}
		if i == 0 && lower == "smid" {
			this.id = lower
		} else if !base.InStrings(lower, keys) {
			var fieldInfo data.FieldInfo
			fieldInfo.Name = v.Name()
			fieldInfo.Type = String2FieldType(v.DatabaseTypeName())
			this.FieldInfos = append(this.FieldInfos, fieldInfo)
		}
	}
}

// "geometry_columns_statistics"  数量和范围
func (this *SqliteFeaset) loadGeoColsSta() {
	sql := `select row_count,extent_min_x,extent_min_y,extent_max_x,extent_max_y 
		from geometry_columns_statistics where f_table_name ="`
	sql += this.Name + "\""
	var vs [5]interface{}
	err := this.store.db.QueryRow(sql).Scan(&vs[0], &vs[1], &vs[2], &vs[3], &vs[4])
	base.PrintError("loadGeoColsSta,"+sql, err)
	needUpdate := false
	if v, ok := vs[0].(int64); ok {
		this.count = v
	} else {
		this.statCount()
		needUpdate = true
	}
	if vs[1] == nil || vs[2] == nil || vs[3] == nil || vs[4] == nil {
		this.Bbox.Init()
		this.statBbox()
		needUpdate = true
	} else {
		this.Bbox.Min.X = vs[1].(float64)
		this.Bbox.Min.Y = vs[2].(float64)
		this.Bbox.Max.X = vs[3].(float64)
		this.Bbox.Max.Y = vs[4].(float64)
	}
	if needUpdate {
		this.updateGeoColsSta()
	}
}

func (this *SqliteFeaset) statCount() {
	// 对象个数好办
	this.store.db.QueryRow("select count(*) from " + this.Name).Scan(&this.count)
}

// 统计数量和范围
func (this *SqliteFeaset) statBbox() {
	// 范围得全部过一遍
	rows, err := this.store.db.Query("select * from " + this.Name)
	base.PrintError("select *", err)
	if err == nil {
		for rows.Next() {
			fea := fetchFea(rows, this.GeoType, this.geom, this.id)
			if fea.Geo != nil {
				this.Bbox = this.Bbox.Union(fea.Geo.GetBounds())
			}
		}
		rows.Close()
	}
}

// 更新 "geometry_columns_statistics"  数量和范围
func (this *SqliteFeaset) updateGeoColsSta() {
	sql := `UPDATE geometry_columns_statistics 
		set row_count=?,extent_min_x=?,extent_min_y=?,extent_max_x=?,extent_max_y=? 
		where f_table_name=?`
	stmt, err := this.store.db.Prepare(sql)
	base.PrintError("update talbe", err)
	stmt.Exec(this.count, this.Bbox.Min.X, this.Bbox.Min.Y, this.Bbox.Max.X, this.Bbox.Max.Y, this.Name)
	defer stmt.Close()
}

// 初始化空间索引
func (this *SqliteFeaset) initSpatailIndex() {
	if this.indexLevel >= 0 {
		this.idx.InitDB(this.Bbox, this.indexLevel)
	} else {
		// 前面读取到的索引层级为负数，说明需要重新构建索引
		this.createSpatailIndex()
	}
}

// 创建空间索引
func (this *SqliteFeaset) createSpatailIndex() {
	// 得到index level
	this.indexLevel = index.CalcZOderLevel(this.count)
	this.idx.InitDB(this.Bbox, this.indexLevel)
	// 创建index_level 字段 store 已经搞定了
	// CreateField(this.store.db, "geometry_columns", "g_index_level", data.TypeInt)
	// CreateField(this.store.db, "geometry_columns", "g_index_type", data.TypeString)
	this.updateIndexFields()

	// 创建 index_code 字段
	CreateField(this.store.db, this.Name, "g_index_code", data.TypeInt)
	// 每条记录，都写入 index_code
	this.updateIndex()
	// 创建数据库索引
	// todo 索引是否会被重复创建？重复了又会如何？
	CreateDbIndex(this.store.db, this.Name, "g_index_code", "g_index_code_"+this.Name)
}

// 更新系统表中的 g_index_level和g_index_type 信息
func (this *SqliteFeaset) updateIndexFields() {
	sql := "UPDATE geometry_columns SET g_index_level=?, g_index_type=? WHERE f_table_name = ?"
	stmt, err := this.store.db.Prepare(sql)
	base.PrintError("update index_level", err)
	stmt.Exec(this.indexLevel, this.idx.Type(), this.Name)
	defer stmt.Close()
}

// 给数据库中的某个字段创建索引，之前要判断该索引是否存在
func CreateDbIndex(db *sql.DB, table, fieldName, indexName string) {
	// CREATE  INDEX index_name on table_name (column_name)
	sql := "create index " + indexName + " on " + table + " (" + fieldName + ")"
	stmt, err := db.Prepare(sql)
	base.PrintError("create db index", err)
	if err == nil {
		stmt.Exec()
	}
	if stmt != nil {
		stmt.Close()
	}
}

// 获取当前rows对应的geometry
// 注意：这里不管Next的事情
func fetchGeo(rows *sql.Rows, geoType geometry.GeoType) geometry.Geometry {
	var id int64
	var geodata []byte
	err := rows.Scan(&id, &geodata)
	if err == nil {
		geo := geometry.CreateGeo(geoType)
		geo.SetID(id)
		geo.From(geodata, geometry.GAIA)
		return geo
	}
	return nil
}

// 读取索引 codes
func (this *SqliteFeaset) loadCodes() (codes map[int]int32) {
	sql := "select " + this.id + "," + this.geom + " from " + this.Name
	rows, err := this.store.db.Query(sql)
	base.PrintError("loadCodes", err)
	codes = make(map[int]int32, this.count)
	if err == nil {
		for rows.Next() {
			geo := fetchGeo(rows, this.GeoType)
			if geo != nil {
				bbox := geo.GetBounds()
				code := this.idx.GetCode(bbox)
				codes[int(geo.GetID())] = code
			}
		}
	}
	defer rows.Close()
	return
}

// 更新 索引code
func (this *SqliteFeaset) updateIndex() {
	// 真要更新索引编码，只能先复制sqlite文件，再一遍查询，一遍更新
	bakName := this.store.filename + ".gbak"
	base.CopyFile(bakName, this.store.filename)

	dbR, err := sql.Open("sqlite3", bakName)
	base.PrintError("sql.Open", err)
	defer dbR.Close()

	sqlSel := "select " + this.id + "," + this.geom + " from " + this.Name
	rs, err := dbR.Query(sqlSel)
	base.PrintError("store.db.Query:"+sqlSel, err)
	defer rs.Close()

	dbW, err := sql.Open("sqlite3", this.store.filename)
	defer dbW.Close()
	base.PrintError("sql.Open", err)
	tx, _ := dbW.Begin()
	update := "UPDATE " + this.Name + " SET g_index_code =? where " + this.id + "=?"
	stmt, _ := tx.Prepare(update)
	defer stmt.Close()

	for rs.Next() {
		geo := fetchGeo(rs, this.GeoType)
		if geo != nil {
			bbox := geo.GetBounds()
			code := this.idx.GetCode(bbox)
			id := geo.GetID()
			_, err := stmt.Exec(code, id)
			base.PrintError("stmt.Exec", err)
		}
	}
	err = tx.Commit()
	base.PrintError("tx.Commit:", err)

	// 最后删除复制的sqlite文件
	err = dbR.Close()
	base.PrintError("dbR.Close:", err)
	err = os.Remove(bakName)
	base.PrintError("os.Remove:", err)
}

// spatialite 中的 geo type转化为 这里的type
func SPL2GeoType(splType int) geometry.GeoType {
	switch splType {
	case 1, 4: // 4估计是多点，先按照单点处理
		return geometry.TGeoPoint
	case 5:
		return geometry.TGeoPolyline
	case 6:
		return geometry.TGeoPolygon
	default:
		fmt.Println("todo")
	}
	return geometry.TGeoEmpty
}

// gogis中的geo type转化为 spatialite中的定义
func GeoType2SPL(geotype geometry.GeoType) int {
	switch geotype {
	case geometry.TGeoPoint: // 4估计是多点，先按照单点处理
		return 1
	case geometry.TGeoPolyline:
		return 5
	case geometry.TGeoPolygon:
		return 6
	default:
		fmt.Println("todo")
	}
	return 0
}

// 创建字段，先查询看看字段是否存在
func CreateField(db *sql.DB, table string, fieldName string, fieldType data.FieldType) {
	if !FieldIsExist(db, table, fieldName) {
		// 真正创建字段
		// ALTER TABLE 表名 ADD COLUMN 列名 数据类型
		sql := "alter table " + table + " add column " + fieldName + " " + FieldType2String(fieldType)
		_, err := db.Exec(sql)
		if err != nil {
			fmt.Print("create field, error:", err)
		}
	}
}

// sql中，字段类型转化为字符串
func FieldType2String(fieldType data.FieldType) string {
	switch fieldType {
	case data.TypeBool:
	case data.TypeInt:
		return "INTEGER"
	case data.TypeFloat:
		return "REAL"
	case data.TypeString:
		return "varchar"
		// todo
		// TypeTime    FieldType = 6 // time.Time TIMESTAMP
	case data.TypeBlob:
		return "BLOB"
	}
	return "Unknown"
}

func String2FieldType(fieldType string) data.FieldType {
	switch strings.ToUpper(fieldType) {
	case "INTEGER":
		return data.TypeInt
	case "REAL":
		return data.TypeFloat
	case "VARCHAR":
		return data.TypeString
		// todo
		// TypeTime    FieldType = 6 // time.Time TIMESTAMP
	case "BLOB":
		return data.TypeBlob
	}
	return data.TypeUnknown
}

// 判断字段是否存在
func FieldIsExist(db *sql.DB, table string, fieldName string) bool {
	rows, err := db.Query("select * from " + table + " limit 0")
	base.PrintError("FieldIsExist", err)
	if rows != nil {
		fields, err := rows.Columns()
		base.PrintError("rows.Columns", err)
		if err == nil {
			for _, v := range fields {
				if v == fieldName {
					return true
				}
			}
		}
		defer rows.Close()
	}
	return false
}

// 几个关键字的字段要剔除掉
func (this *SqliteFeaset) exKeyFields() {
	fields := []string{this.id, this.geom, "g_index_level"}
	for i, v := range this.FieldInfos {
		if base.InStrings(strings.ToLower(v.Name), fields) {
			this.FieldInfos = append(this.FieldInfos[0:i], this.FieldInfos[i+1:]...)
		}
	}
}

func (this *SqliteFeaset) Close() {
	this.idx.Clear()
}

func (this *SqliteFeaset) GetStore() data.Datastore {
	return this.store
}

func (this *SqliteFeaset) GetCount() int64 {
	return this.count
}

func (this *SqliteFeaset) BeforeWrite(count int64) {
	this.indexLevel = index.CalcZOderLevel(count)
	this.idx.InitDB(this.Bbox, this.indexLevel)
	this.updateIndexFields()
}

// 批量写入数据
func (this *SqliteFeaset) BatchWrite(feas []data.Feature) {
	// 	INSERT INTO TABLE_NAME [(column1, column2, column3,...columnN)]
	// VALUES (value1, value2, value3,...valueN);
	sqlInsert := "INSERT INTO " + this.Name + "(" + this.geom + ",g_index_code"
	for _, v := range this.FieldInfos {
		sqlInsert += "," + v.Name
	}
	sqlInsert += ")  values("
	fieldCount := len(this.FieldInfos)
	for i := 0; i < fieldCount+1; i++ {
		sqlInsert += "?,"
	}
	sqlInsert += "?)"

	// db, err := sql.Open("sqlite3", this.store.filename)
	// base.PrintError("sql.Open", err)
	// defer db.Close()
	tx, err := this.store.db.Begin()
	base.PrintError("db.Begin", err)

	stmt, err := tx.Prepare(sqlInsert)
	defer stmt.Close()
	base.PrintError("insert Prepare", err)
	for _, v := range feas {
		vs := make([]interface{}, fieldCount+2)
		vs[0] = v.Geo.To(geometry.GAIA)
		vs[1] = this.idx.GetCode(v.Geo.GetBounds())
		for j, info := range this.FieldInfos {
			vs[j+2] = v.Atts[info.Name]
		}

		_, err := stmt.Exec(vs...)
		base.PrintError("insert one recode", err)
		// this.Bbox = this.Bbox.Union(v.Geo.GetBounds())
	}
	tx.Commit()

	this.lock.Lock()
	this.count += int64(len(feas))
	this.lock.Unlock()
}

func (this *SqliteFeaset) EndWrite() {
	// 更新 geometry_columns 表
	this.updateGeoColsSta()
}

// 构造sql中的in语句
func buildSqlIn(values []int32) string {
	in := " IN("
	count := len(values)
	for i, v := range values {
		in += strconv.Itoa(int(v))
		if i != count-1 {
			in += ","
		}
	}
	in += ")"
	return in
}

// 综合查询
func (this *SqliteFeaset) Query(def *data.QueryDef) data.FeatureIterator {
	if def == nil {
		def = data.NewQueryDef(this.Bbox)
	}
	feaitr := new(SqliteFeaItr)
	feaitr.feaset = this
	feaitr.fields = def.Fields

	// 根据空间查询条件做筛选
	feaitr.squery.Init(def.SpatialObj, def.SpatialMode)
	feaitr.codes = feaitr.squery.QueryCodes(this.idx)
	codes := feaitr.codes
	if float64(len(feaitr.codes)*1.0/index.CalcCodeCount(int(this.indexLevel))) > 0.8 {
		codes = nil // 八成都要，就不做codes过滤了
	}
	feaitr.where = def.Where
	feaitr.geotype = GeoType2SPL(this.GeoType)
	from := " from " + this.Name
	where := buildWhere(codes, def.Where)
	// fmt.Println("where:", where)
	if len(strings.TrimSpace(where)) != 0 {
		err := this.store.db.QueryRow("select count(*) " + from + where).Scan(&feaitr.count)
		base.PrintError("QueryRow", err)
	} else {
		feaitr.count = this.count
	}

	return feaitr
}
