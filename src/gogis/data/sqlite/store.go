// 基于sqlite实现的空间文件引擎，与spatialite兼容
package sqlite

import (
	"database/sql"
	"gogis/base"
	"gogis/data"
	"gogis/index"

	// "database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	data.RegisterDatastore(data.StoreSqlite, NewSqliteStore)
}

func NewSqliteStore() data.Datastore {
	return new(SqliteStore)
}

// sqlite 数据存储库，采用 spatialite 空间存储
type SqliteStore struct {
	db               *sql.DB
	filename         string
	data.Featuresets // 匿名组合
	params           data.ConnParams
}

// 打开sqlite文件
// 通过 ConnParams["filename"] 输入文件名，不存在时自动创建
func (this *SqliteStore) Open(params data.ConnParams) (res bool, err error) {
	this.filename = params["filename"].(string)
	this.params = params
	// 首先得看文件在不在，来确定是打开，还是创建
	exist := base.IsExist(this.filename)
	if exist {
		this.open()
	} else {
		this.create()
	}
	return true, nil
}

func (this *SqliteStore) open() {
	this.db, _ = sql.Open("sqlite3", this.filename)
	// 读取系统表
	this.loadGeoColumns()
}

// 确保有 index_level 字段
func (this *SqliteStore) ensureIndexFields() {
	this.ensureSysField("g_index_level", data.TypeInt)
	this.ensureSysField("g_index_type", data.TypeString)
}

func (this *SqliteStore) ensureSysField(field string, ftype data.FieldType) {
	sql := "select " + field + " from geometry_columns"
	rows, err := this.db.Query(sql)
	if err != nil {
		CreateField(this.db, "geometry_columns", field, ftype)
	} else if rows != nil {
		defer rows.Close()
	}
}

// 加载系统表
func (this *SqliteStore) loadGeoColumns() {
	this.ensureIndexFields()

	// 再读取 名字、geom字段、类型和投影系统
	sql := `select f_table_name,f_geometry_column,geometry_type,srid, 
			g_index_level, g_index_type from geometry_columns`
	rows, err := this.db.Query(sql)
	base.PrintError("loadGeoColumns", err)
	for i := 0; rows.Next(); i++ {
		var name, geom string
		var geotype, srid int
		var indexLevel, indextype interface{}
		err := rows.Scan(&name, &geom, &geotype, &srid, &indexLevel, &indextype)
		base.PrintError("load feasets", err)
		if err == nil {
			feaset := new(SqliteFeaset)
			feaset.store = this
			feaset.Name = name
			feaset.geom = geom
			feaset.GeoType = SPL2GeoType(geotype)
			feaset.Proj = base.PrjFromEpsg(srid)

			if level, ok := indexLevel.(int64); ok {
				feaset.indexLevel = int32(level)
			} else {
				feaset.indexLevel = -1
			}
			feaset.idx = index.NewSpatialIndexDB(this.getIndexType(indextype))

			this.Feasets = append(this.Feasets, feaset)
		}
	}
	defer rows.Close() // 记得关闭
}

// 确定空间索引类型
func (this *SqliteStore) getIndexType(db interface{}) (indextype index.SpatialIndexType) {
	indextype = index.TypeZOrderIndex // 默认用xz-order
	if itypeInParams := this.params.GetString("index"); len(itypeInParams) > 0 {
		// 用户指定了，先用指定的
		indextype = index.SpatialIndexType(itypeInParams)
	} else if itypeInDB, ok := db.(string); ok && len(itypeInDB) > 0 {
		// 没有指定，而数据库中存储了，就用存储的
		indextype = index.SpatialIndexType(itypeInDB)
	}
	return
}

// 创建存储库
func (this *SqliteStore) create() {
	// 先复制 文件
	src := base.GlobalProj().Filename
	base.CopyFile(this.filename, src)
	this.db, _ = sql.Open("sqlite3", this.filename)

	// 创建系统表
	// this.createGeoColumns()
	// geometry_columns_auth
	// this.createGeoColsAuth()
	// 创建ref表，并写入epsg内容
	// this.createSpatialRef()
	// this.writeSpatialRef()
	// 创建字段表
	// this.createGeoColsField()
	// 创建统计表
	// this.createGeoColsSta()
}

// 创建字段表
func (this *SqliteStore) createGeoColsField() {
	sql := `CREATE TABLE geometry_columns_field_infos (
		f_table_name varchar not null,
		f_geometry_column varchar,
		ordinal INTEGER,
		column_name varchar )`
	stmt, err := this.db.Prepare(sql)
	base.PrintError("createGeoColsField", err)
	stmt.Exec()
	defer stmt.Close()
}

// 创建统计表
func (this *SqliteStore) createGeoColsSta() {
	sql := `CREATE TABLE geometry_columns_statistics (
		f_table_name varchar not null,
		f_geometry_column varchar,
		row_count INTEGER,
		extent_min_x real,
		extent_min_y real,
		extent_max_x real,
		extent_max_y real )`
	stmt, err := this.db.Prepare(sql)
	base.PrintError("createGeoColsSta", err)
	stmt.Exec()
	defer stmt.Close()
}

// 创建系统表
func (this *SqliteStore) createGeoColumns() {
	createSql := `CREATE TABLE geometry_columns (
		f_table_name varchar not null,
		f_geometry_column varchar,
		geometry_type INTEGER,
		srid INTEGER,
		g_index_level INTEGER)`
	stmt, err := this.db.Prepare(createSql)
	base.PrintError("createGeoColumns", err)
	stmt.Exec()
	defer stmt.Close()
}

func (this *SqliteStore) createGeoColsAuth() {
	createSql := `CREATE TABLE geometry_columns_auth (
		f_table_name varchar not null,
		f_geometry_column varchar,
		read_only INTEGER,
		hidden INTEGER)`
	stmt, err := this.db.Prepare(createSql)
	base.PrintError("createGeoColsAuth", err)
	stmt.Exec()
	defer stmt.Close()
}

// func (this *SqliteStore) createSpatialRef() {
// 	createSql := `CREATE TABLE spatial_ref_sys (
// 		srid INTEGER,
// 		auth_name varchar,
// 		auth_srid INTEGER,
// 		ref_sys_name varchar,
// 		proj4text varchar,
// 		srtext varchar
// 	  )`
// 	stmt, err := this.db.Prepare(createSql)
// 	base.PrintError("create table", err)
// 	stmt.Exec()
// 	defer stmt.Close()
// }

// func (this *SqliteStore) writeSpatialRef() {
// 	proj := base.GlobalProj()
// sql := `INSERT INTO spatial_ref_sys
// (srid,auth_name,auth_srid,ref_sys_name,proj4text,srtext)
// VALUES (?, ?, ?,?,?,?)`
// stmt, _ := this.db.Prepare(sql)
// defer stmt.Close()

// proj := base.GlobalProj()
// if proj != nil {
// 	for _, v := range proj.ProjInfos {
// 		stmt.Exec(v.Epsg, "epsg", v.Epsg, v.Name, v.Proj4, v.Wkt)
// 	}
// }
// }

func (this *SqliteStore) GetType() data.StoreType { // 得到存储类型
	return data.StoreSqlite
}

func (this *SqliteStore) GetConnParams() data.ConnParams {
	params := data.NewConnParams()
	params["filename"] = this.filename
	params["type"] = string(this.GetType())
	params["gowrite"] = 1
	return params
}

// 创建数据集，返回创建好的对象
func (this *SqliteStore) CreateFeaset(info data.FeasetInfo) data.Featureset {
	// 先保证name不重复
	info.Name = base.GetUniqueName(info.Name, this.GetFeasetNames())
	feaset := new(SqliteFeaset)
	feaset.FeasetInfo = info
	feaset.id = "pk"
	feaset.geom = "geom"
	feaset.store = this
	feaset.exKeyFields() // 剔除保留的字段关键字

	// 开启事务
	tx, err := this.db.Begin()
	base.PrintError("db.Begin", err)

	// 在geometry_columns表中插入一条记录
	{
		insertSql := `insert into geometry_columns
		(f_table_name,f_geometry_column,geometry_type,coord_dimension,srid, 
		spatial_index_enabled,g_index_level,g_index_type) 
		values (?,?,?,?,?,?,?)`
		stmt, err := tx.Prepare(insertSql)
		base.PrintError("insert geometry_columns", err)
		epsg := 0
		if info.Proj != nil {
			epsg = info.Proj.Epsg
		}
		indextype := this.params.GetString("index")
		if len(indextype) == 0 {
			indextype = "xzorder"
		}
		feaset.idx = index.NewSpatialIndexDB(index.SpatialIndexType(indextype))
		_, err = stmt.Exec(info.Name, "geom", GeoType2SPL(info.GeoType), 2, epsg, 0, -1, indextype)
		base.PrintError("stmt.Exec", err)
		defer stmt.Close()
	}

	// geometry_columns_auth
	// {
	// 	insertSql := `insert into geometry_columns_auth
	// 	(f_table_name,f_geometry_column,read_only,hidden)
	// 	values (?,?,?,?)`
	// 	stmt, err := tx.Prepare(insertSql)
	// 	base.PrintError("insert geometry_columns_auth", err)
	// 	_, err = stmt.Exec(info.Name, "geom", 0, 0)
	// 	base.PrintError("stmt.Exec", err)
	// 	defer stmt.Close()
	// }

	// 在 geometry_columns_field_infos中插入记录
	{ // (f_table_name,f_geometry_column,ordinal,column_name)
		insertSql := `insert into geometry_columns_field_infos 
		values (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
		stmt, err := tx.Prepare(insertSql)
		base.PrintError("insert geometry_columns_field_infos", err)

		_, err = stmt.Exec(info.Name, feaset.geom, 0, "pk", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		base.PrintError("stmt.Exec", err)
		_, err = stmt.Exec(info.Name, feaset.geom, 1, feaset.geom, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		base.PrintError("stmt.Exec", err)
		_, err = stmt.Exec(info.Name, feaset.geom, 2, "g_index_code", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		base.PrintError("stmt.Exec", err)
		for i := 0; i < len(feaset.FieldInfos); i++ {
			_, err = stmt.Exec(info.Name, feaset.geom, i+3, feaset.FieldInfos[i].Name, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
			base.PrintError("stmt.Exec", err)
		}
		defer stmt.Close()
	}

	{ // 在 geometry_columns_statistics 中插入记录
		insertSql := `insert into geometry_columns_statistics
		(f_table_name,f_geometry_column,row_count,extent_min_x,extent_min_y, extent_max_x, extent_max_y) 
		values (?,?,?,?,?,?,?)`
		stmt, err := tx.Prepare(insertSql)
		base.PrintError("insert geometry_columns_statistics", err)
		stmt.Exec(info.Name, feaset.geom, 0, nil, nil, nil, nil)
		defer stmt.Close()
	}

	{ // 创建一张要素表
		createSql := "CREATE TABLE " + feaset.Name
		createSql += "(pk INTEGER PRIMARY KEY AUTOINCREMENT,geom blob,g_index_code INTEGER"
		for _, v := range feaset.FieldInfos {
			createSql += ", " + v.Name + " " + FieldType2String(v.Type)
		}
		createSql += ")"
		stmt, err := tx.Prepare(createSql)
		base.PrintError("CREATE TABLE", err)
		stmt.Exec()
		defer stmt.Close()
	}

	{ // 创建空间索引字段的数据库索引 CREATE INDEX index_name ON table_name (column_name);
		indexName := "g_index_code_" + feaset.Name
		sql := "CREATE INDEX " + indexName + " ON " + feaset.Name + " (g_index_code)"
		stmt, err := tx.Prepare(sql)
		base.PrintError("CREATE INDEX", err)
		if err == nil {
			stmt.Exec()
		}
		if stmt != nil {
			stmt.Close()
		}
	}

	tx.Commit() // 提交事务
	return feaset
}

func (this *SqliteStore) DeleteFeaset(name string) bool {
	feaset, n := this.GetFeasetByName(name)
	if feaset != nil {
		// 开启事务
		tx, _ := this.db.Begin()

		// 删除系统表中的相关 记录
		tables := []string{"geometry_columns", "geometry_columns_auth", "geometry_columns_field_infos", "geometry_columns_statistics"}
		for _, v := range tables {
			sql := `delete from ? where f_table_name=?`
			stmt, err := tx.Prepare(sql)
			base.PrintError("delete geometry_columns", err)
			stmt.Exec(v, feaset.GetName())
			defer stmt.Close()
		}
		// 删除要素表 DROP TABLE COMPANY
		{
			sql := "DROP TABLE ?"
			stmt, err := tx.Prepare(sql)
			base.PrintError("DROP TABLE", err)
			stmt.Exec(feaset.GetName())
			defer stmt.Close()
		}

		tx.Commit() // 提交事务
		feaset.Close()
		this.Feasets = append(this.Feasets[:n], this.Feasets[n+1:]...)
		return true
	}
	return false
}

// 关闭，释放资源
func (this *SqliteStore) Close() {
	this.db.Close()
	for _, v := range this.Feasets {
		v.Close()
	}
	this.Feasets = this.Feasets[:0]
}
