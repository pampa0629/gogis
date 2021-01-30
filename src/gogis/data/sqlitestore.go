// 基于sqlite实现的空间文件引擎
package data

import (
	"database/sql"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"gogis/index"
	"strconv"

	// "database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	RegisterDatastore(StoreSqlite, NewSqliteStore)
}

func NewSqliteStore() Datastore {
	return new(SqliteStore)
}

// sqlite 数据存储库，采用 spatialite 空间存储
type SqliteStore struct {
	db       *sql.DB
	filename string
	// feasets  []*SqliteFeaset
	Feasets // 匿名组合
}

// sst: spatial sys table
const SST_GEO_COLS = "geometry_columns" // f_table_name,f_geometry_column,geometry_type,......

// 打开sqlite文件
// 通过 ConnParams["filename"] 输入文件名，不存在时自动创建
func (this *SqliteStore) Open(params ConnParams) (res bool, err error) {
	this.filename = params["filename"].(string)
	this.db, err = sql.Open("sqlite3", this.filename)

	// 读取系统表
	// 先知道数量
	var count int64
	this.db.QueryRow("select count(*) from " + SST_GEO_COLS).Scan(&count)
	this.feasets = make([]Featureset, count)
	// 再读取 名字、geom字段、类型和投影系统
	rows, err := this.db.Query("select f_table_name,f_geometry_column,geometry_type,srid from " + SST_GEO_COLS)
	if err == nil {
		this.loadSys(rows)
	} else {
		// 没有就创建系统表 todo
		// this.createSys()
	}
	return true, nil
}

// spatialite table 3's elements
type st3 struct {
	name    string
	geom    string
	geotype int
	srid    int
}

// 加载系统表
func (this *SqliteStore) loadSys(rows *sql.Rows) {
	st3s := make([]st3, 0)
	for i := 0; rows.Next(); i++ {
		var st st3
		err := rows.Scan(&st.name, &st.geom, &st.geotype, &st.srid)
		if err == nil {
			st3s = append(st3s, st)
		}
	}
	rows.Close() // 记得关闭

	for i, v := range st3s {
		feaset := new(SqliteFeaset)
		feaset.store = this
		feaset.name = v.name
		feaset.geom = v.geom
		feaset.geotype = v.geotype
		feaset.proj = base.PrjFromEpsg(v.srid)
		this.feasets[i] = feaset
	}

}

func (this *SqliteStore) CreateFeaset(name string, bbox base.Rect2D, geotype geometry.GeoType) Featureset {
	return nil
}

// 创建系统表
// func (this *SqliteStore) createSys() {
// 	createSql := `CREATE TABLE $1(
// 		Name varchar not null,
// 		MinX decimal,
// 		MinY decimal,
// 		MaxX decimal,
// 		MaxY decimal,
// 		IndexType INTEGER
// 	  )`
// 	stmt, err := this.db.Prepare(createSql)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer stmt.Close()

// 	stmt.Exec(SYS_TABLE_NAME)
// 	// result, err := stmt.Exec(SYS_TABLE_NAME)
// }

func (this *SqliteStore) GetType() StoreType { // 得到存储类型
	return StoreSqlite
}

func (this *SqliteStore) GetConnParams() ConnParams {
	params := NewConnParams()
	params["filename"] = this.filename
	params["type"] = string(this.GetType())
	return params
}

// 创建数据集，返回创建好的对象
// func (this *SqliteStore) CreateFeaset(name string, geoType geometry.GeoType, finfos []FieldInfo) (Featureset, error) {
// 	// 先保证name不重复
// 	for _, v := range this.feasets {
// 		if v.name == name {
// 			return v, errors.New("feature set of name: " + name + " have existed.")
// 		}
// 	}

// 	// 在系统表中插入一条记录
// 	insertSql := "insert into " + SYS_TABLE_NAME + "(Name)" + " values ($1)"
// 	stmt, err := this.db.Prepare(insertSql)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer stmt.Close()
// 	stmt.Exec(name)

// 	// 创建一张新表
// 	createSql := `CREATE TABLE $1(
// 		gid INTEGER,
// 		geom data,
// 		gCode INTEGER
// 	  )`
// 	stmt, err = this.db.Prepare(createSql)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer stmt.Close()
// 	stmt.Exec(name)

// 	// todo 返回新创建的数据集
// 	return nil, nil
// }

// func (this *SqliteStore) GetFeasetByNum(num int) (Featureset, error) {
// 	if num >= 0 && num < len(this.feasets) {
// 		return this.feasets[num], nil
// 	}
// 	return nil, errors.New("num must big than zero and less the count of feature sets.")
// }

// func (this *SqliteStore) GetFeasetByName(name string) (Featureset, error) {
// 	for _, v := range this.feasets {
// 		if strings.ToLower(v.name) == strings.ToLower(name) {
// 			return v, nil
// 		}
// 	}
// 	return nil, errors.New("cannot find the feature set of name: " + name + ".")
// }

// func (this *SqliteStore) GetFeasetNames() (names []string) {
// 	names = make([]string, len(this.feasets))
// 	for i, _ := range names {
// 		names[i] = this.feasets[i].name
// 	}
// 	return
// }

// 关闭，释放资源
func (this *SqliteStore) Close() {
	this.db.Close()
	for _, v := range this.feasets {
		v.Close()
	}
	this.feasets = this.feasets[:0]
}

// 矢量数据集合
type SqliteFeaset struct {
	name    string
	geom    string // geometry用的字段，默认为"geom"
	geotype int    // gaia 的类型
	bbox    base.Rect2D
	count   int64 // 对象个数
	// indexType index.SpatialIndexType
	idx index.ZOrderIndex // 暂时只支持 zorder索引
	// idx index.SpatialIndexDB
	// index data 先不加载
	projCommon
	store *SqliteStore
}

// sst: spatial sys table
const SST_GEO_STA = "geometry_columns_statistics"

func (this *SqliteFeaset) Open() (bool, error) {
	sql := "select row_count,extent_min_x,extent_min_y,extent_max_x,extent_max_y from " + SST_GEO_STA
	sql += " where f_table_name = '" + this.name + "'"
	// fmt.Println(sql)
	err := this.store.db.QueryRow(sql).Scan(&this.count, &this.bbox.Min.X, &this.bbox.Min.Y, &this.bbox.Max.X, &this.bbox.Max.Y)
	if err == nil {
		// 再判断 空间索引是否存在，不存在 则要创建之
		this.loadSpatailIndex()
		return true, nil
	}

	return false, err
}

// ssf: spatial sys field
const SSF_INDEX_LEVEL = "index_level"

// 读取/创建空间索引
func (this *SqliteFeaset) loadSpatailIndex() {
	sql := "select " + SSF_INDEX_LEVEL + " from " + SST_GEO_STA
	sql += " where f_table_name = '" + this.name + "'"
	var indexLevel int32
	err := this.store.db.QueryRow(sql).Scan(&indexLevel)
	// 索引列已经存在，继续读取索引
	if err == nil && indexLevel >= 0 { // < 0 表示索引失效了
		this.idx.InitDB(this.bbox, indexLevel)
		return
	}
	// 但凡前面读取索引出问题，就重新构建索引
	this.createSpatailIndex()
}

// ssf: spatial sys field 要素表中，code字段
const SSF_INDEX_CODE = "index_code"

// ssf: spatial sys field 要素表中，code字段的数据库索引
const SSF_DB_INDEX_CODE = "db_index_code"

// 创建空间索引
func (this *SqliteFeaset) createSpatailIndex() {
	// 1，写入level
	// 1.1 得到  level
	level := index.CalcZOderLevel(this.count)
	// level = 5 // todo
	this.idx.InitDB(this.bbox, level)
	// 1.2 创建字段
	CreateField(this.store.db, SST_GEO_STA, SSF_INDEX_LEVEL, TypeInt)
	// 1.3 写入 level
	// UPDATE COMPANY SET ADDRESS = 'Texas' WHERE ID = 6
	sql := "UPDATE " + SST_GEO_STA + " SET " + SSF_INDEX_LEVEL + " = " + strconv.Itoa(int(level))
	sql += " WHERE " + " f_table_name = '" + this.name + "'"
	_, err := this.store.db.Exec(sql)
	if err != nil {
		fmt.Println("update level, err:", err)
	}

	// 2，创建 index_code 字段
	CreateField(this.store.db, this.name, SSF_INDEX_CODE, TypeInt)

	// 3，每条记录，都写入 code 内容
	this.updateIndex()

	// 2.1 创建 数据库索引
	CreateDbIndex(this.store.db, this.name, SSF_INDEX_CODE, SSF_DB_INDEX_CODE)
}

// 给数据库中的某个字段创建索引，之前要判断该索引是否存在
func CreateDbIndex(db *sql.DB, table, fieldName, indexName string) {
	// CREATE  INDEX index_name on table_name (column_name)
	sql := "create index " + indexName + " on " + table + " (" + fieldName + ")"
	_, err := db.Exec(sql)
	if err != nil {
		fmt.Println("create db index error:", err)
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
	sql := "select pk," + this.geom + " from " + this.name
	rows, err := this.store.db.Query(sql)
	codes = make(map[int]int32, this.count)
	if err == nil {
		for rows.Next() {
			geo := fetchGeo(rows, SPL2GeoType(this.geotype))
			if geo != nil {
				bbox := geo.GetBounds()
				code := this.idx.GetCode(bbox)
				codes[int(geo.GetID())] = code
			}
		}
	}
	rows.Close()
	return
}

// 更新 索引code
func (this *SqliteFeaset) updateIndex() {
	codes := this.loadCodes()

	db, err := sql.Open("sqlite3", this.store.filename)
	if err == nil {
		tx, _ := db.Begin()
		update := "UPDATE " + this.name + " SET " + SSF_INDEX_CODE + " =? where pk=?"
		stmt, _ := tx.Prepare(update)
		defer stmt.Close()

		for i, v := range codes {
			_, err := stmt.Exec(v, i)
			if err != nil {
				fmt.Println("update  code, error:", err)
			}
		}
		tx.Commit()
	}
	db.Close()
}

// spatialite 中的 geo type转化为 这里的type
func SPL2GeoType(splType int) geometry.GeoType {
	switch splType {
	case 1, 4: // 4估计是多点，先按照单点处理
		return geometry.TGeoPoint
	case 6:
		return geometry.TGeoPolygon
	}
	return geometry.TGeoEmpty
}

// 创建字段，先查询看看字段是否存在
func CreateField(db *sql.DB, table string, fieldName string, fieldType FieldType) {
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
func FieldType2String(fieldType FieldType) string {
	switch fieldType {
	case TypeBool:
	case TypeInt:
		return "INTEGER"
	case TypeFloat:
		return "REAL"
	case TypeString:
		return "TEXT"
		// todo
		// TypeTime    FieldType = 6 // time.Time TIMESTAMP
	case TypeBlob:
		return "BLOB"
	}
	return "Unknown"
}

// 判断字段是否存在
func FieldIsExist(db *sql.DB, table string, fieldName string) bool {
	rows, err := db.Query("select * from " + table + " limit 0")
	if err == nil {
		fields, err := rows.Columns()
		if err == nil {
			for _, v := range fields {
				if v == fieldName {
					rows.Close()
					return true
				}
			}
		}
	}
	rows.Close()
	return false
}

func (this *SqliteFeaset) GetGeoType() geometry.GeoType {
	return SPL2GeoType(this.geotype)
}

func (this *SqliteFeaset) Close() {
	// todo 这里干点什么呢？
}

func (this *SqliteFeaset) GetStore() Datastore {
	return this.store
}

func (this *SqliteFeaset) GetName() string {
	return this.name
}

func (this *SqliteFeaset) GetCount() (count int64) { // 对象个数
	this.store.db.QueryRow("select count(*) from " + this.name).Scan(&count)
	return count
}

func (this *SqliteFeaset) GetBounds() base.Rect2D {
	return this.bbox
}

func (this *SqliteFeaset) GetFieldInfos() (finfos []FieldInfo) {
	// open 时，应加载这些
	// select * from table 就能搞定
	return
}

// 批量写入数据 todo
func (this *SqliteFeaset) BatchWrite(feas []Feature) {
}

func (this *SqliteFeaset) EndWrite() {
}

// func (this *SqliteFeaset) Query(bbox base.Rect2D, def QueryDef) FeatureIterator {
// 	return nil
// }

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
func (this *SqliteFeaset) QueryByDef(def QueryDef) FeatureIterator {
	feaitr := new(SqliteFeaItr)
	feaitr.feaset = this
	feaitr.fields = def.Fields

	// 根据空间查询条件做筛选
	feaitr.squery.Init(def.SpatialObj, def.SpatialMode)
	feaitr.codes = feaitr.squery.QueryCodes(&this.idx)
	feaitr.where = def.Where
	feaitr.geotype = this.geotype
	from := " from " + this.name
	where := buildWhere(feaitr.codes, def.Where)
	this.store.db.QueryRow("select count(*) " + from + where).Scan(&feaitr.count)
	return feaitr
}

func (this *SqliteFeaset) QueryByBounds(bbox base.Rect2D) FeatureIterator {
	var def QueryDef
	def.SpatialMode = base.Intersects
	def.SpatialObj = bbox
	return this.QueryByDef(def)
	// codes := this.idx.QueryDB(bbox)

	// from := " from " + this.name

	// // where := " where " + SSF_INDEX_CODE
	// // in := buildSqlIn(codes)
	// where := buildWhere(codes, "")
	// itr := new(SqliteFeaItr)
	// itr.feaset = this
	// itr.bbox = bbox
	// itr.codes = codes
	// itr.geotype = this.geotype
	// this.store.db.QueryRow("select count(*) " + from + where).Scan(&itr.count)
	// return itr
}

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
	squery  SpatailQuery

	countPerGo int // 每一个批次的对象数量
}

func (this *SqliteFeaItr) Count() int64 {
	return this.count
}

func (this *SqliteFeaItr) Close() {
	return
}

// todo
func (this *SqliteFeaItr) Next() (fea Feature, ok bool) {
	return
}

// 为了批量读取做准备，返回批量的次数
func (this *SqliteFeaItr) PrepareBatch(objCount int) (goCount int) {
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
	// if len(codes) > 0 {
	if codes != nil {
		// 如果编码存在，且长度为0，则说明啥也不能要
		if len(codes) == 0 {
			codes = []int32{-1}
		}
		in = SSF_INDEX_CODE + buildSqlIn(codes)
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
	if len(this.fields) > 0 {
		sel += " pk, " + this.feaset.geom
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
func (this *SqliteFeaItr) BatchNext(batchNo int) (feas []Feature, result bool) {
	// fmt.Println("db open:" + this.feaset.store.filename)
	db, err := sql.Open("sqlite3", this.feaset.store.filename)
	if err == nil && db != nil {
		sel := this.buildSelect()
		from := " from " + this.feaset.name
		codes := []int32{}
		if this.codess != nil && batchNo < len(this.codess) {
			codes = this.codess[batchNo]
		}
		where := buildWhere(codes, this.where)
		sql := sel + from + where
		rows, err := db.Query(sql)
		if err == nil {
			result = true
			geotype := SPL2GeoType(this.geotype)
			feas = make([]Feature, 0, this.countPerGo)
			for rows.Next() {
				fea := fetchFea(rows, geotype, this.feaset.geom)
				if fea.Geo != nil && this.squery.Match(fea.Geo) {
					feas = append(feas, fea)
				}
			}
			rows.Close()
		} else {
			fmt.Println("db query, sql:", sql, " error:", err)
		}
		defer db.Close()
	} else {
		fmt.Println("db open:"+this.feaset.store.filename+" error:", err)
	}
	// fmt.Println("sqlite batch next, codes:", this.codess[batchNo], " count:", len(feas))
	return
}

// 获取feature
func fetchFea(rows *sql.Rows, geoType geometry.GeoType, geom string) (fea Feature) {
	cols, _ := rows.Columns()
	atts := make([]interface{}, len(cols))
	for i, _ := range atts {
		atts[i] = new(interface{})
	}
	err := rows.Scan(atts...)
	if err == nil {
		fea.Geo = geometry.CreateGeo(geoType)
		fea.Atts = make(map[string]interface{})
		for i, v := range cols {
			att := *atts[i].(*interface{})
			switch v {
			case "pk":
				fea.Geo.SetID(att.(int64))
			case geom:
				fea.Geo.From(att.([]byte), geometry.GAIA)
			default:
				fea.Atts[v] = att
			}
		}
	} else {
		fmt.Println("fetch feature error:", err)
	}
	return
}
