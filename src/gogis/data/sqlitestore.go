// 基于sqlite实现的空间文件引擎
package data

import (
	"database/sql"
	"errors"
	"gogis/base"
	"gogis/geometry"
	"gogis/index"
	"log"

	// "database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// sqlite数据存储库
type SqliteStore struct {
	db       *sql.DB
	filename string
	feasets  []*SqliteFeaset
}

const SYS_TABLE_NAME = "gogis_sys"

// 打开sqlite文件
// 通过 ConnParams["filename"] 输入文件名，不存在时自动创建
func (this *SqliteStore) Open(params ConnParams) (res bool, err error) {
	this.db, err = sql.Open("sqlite3", params["filename"])

	// 读取系统表
	var count int64
	this.db.QueryRow("select count(*) from " + SYS_TABLE_NAME).Scan(&count)
	this.feasets = make([]*SqliteFeaset, count)
	// todo 修改bbox
	rows, err := this.db.Query("select Name,Bbox,indexType from " + SYS_TABLE_NAME)
	if err == nil {
		this.loadSys(rows)
	} else {
		// 没有就创建系统表
		this.createSys()
	}
	return true, nil
}

// 加载系统表
func (this *SqliteStore) loadSys(rows *sql.Rows) {
	var name string
	var bbox base.Rect2D
	var indexType int
	// var indexData []byte
	for i := 0; rows.Next(); i++ {
		err := rows.Scan(&name, &bbox, &indexType)
		if err != nil {
			this.feasets[i] = new(SqliteFeaset)
			this.feasets[i].store = this
			this.feasets[i].bbox = bbox // todo 修改
			this.feasets[i].indexType = index.SpatialIndexType(indexType)
		}
	}
}

// 创建系统表
func (this *SqliteStore) createSys() {
	createSql := `CREATE TABLE $1(
		Name varchar not null,
		MinX decimal,
		MinY decimal,
		MaxX decimal,
		MaxY decimal,
		IndexType INTEGER 
	  )`
	stmt, err := this.db.Prepare(createSql)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	stmt.Exec(SYS_TABLE_NAME)
	// result, err := stmt.Exec(SYS_TABLE_NAME)
}

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
func (this *SqliteStore) CreateFeaset(name string, geoType geometry.GeoType, finfos []FieldInfo) (Featureset, error) {
	// 先保证name不重复
	for _, v := range this.feasets {
		if v.name == name {
			return v, errors.New("feature set of name: " + name + " have existed.")
		}
	}

	// 在系统表中插入一条记录
	insertSql := "insert into " + SYS_TABLE_NAME + "(Name)" + " values ($1)"
	stmt, err := this.db.Prepare(insertSql)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	stmt.Exec(name)

	// 创建一张新表
	createSql := `CREATE TABLE $1(
		gid INTEGER,
		geom data,
		gCode INTEGER
	  )`
	stmt, err = this.db.Prepare(createSql)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	stmt.Exec(name)

	// todo 返回新创建的数据集
	return nil, nil
}

func (this *SqliteStore) GetFeasetByNum(num int) (Featureset, error) {
	if num >= 0 && num < len(this.feasets) {
		return this.feasets[num], nil
	}
	return nil, errors.New("num must big than zero and less the count of feature sets.")
}

func (this *SqliteStore) GetFeasetByName(name string) (Featureset, error) {
	for _, v := range this.feasets {
		if v.name == name {
			return v, nil
		}
	}
	return nil, errors.New("cannot find the feature set of name: " + name + ".")
}

func (this *SqliteStore) FeaturesetNames() (names []string) {
	names = make([]string, len(this.feasets))
	for i, _ := range names {
		names[i] = this.feasets[i].name
	}
	return
}

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
	name      string
	bbox      base.Rect2D
	indexType index.SpatialIndexType
	// index data 先不加载
	store *SqliteStore
}

func (this *SqliteFeaset) Open(name string) (bool, error) {
	// todo 这里干点什么呢？
	// 加载索引？
	return false, nil
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

func (this *SqliteFeaset) Count() (count int64) { // 对象个数
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

func (this *SqliteFeaset) Query(bbox base.Rect2D, def QueryDef) FeatureIterator {
	return nil
}

func (this *SqliteFeaset) QueryByBounds(bbox base.Rect2D) FeatureIterator {
	return nil
}

func (this *SqliteFeaset) QueryByDef(def QueryDef) FeatureIterator {
	return nil
}

// 迭代器
type SqliteFeaItr struct {
}

func (this *SqliteFeaItr) Count() int64 {
	return 0
}

func (this *SqliteFeaItr) Close() {
	return
}

func (this *SqliteFeaItr) Next() (fea Feature, ok bool) {
	return
}

// 只要读取到一个数据，达不到count的要求，也返回true
func (this *SqliteFeaItr) BatchNext(count int) ([]Feature, bool) {
	return nil, false
}
