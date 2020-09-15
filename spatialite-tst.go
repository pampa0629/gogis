package main

import (
	"database/sql"
	"fmt"

	_ "github.com/briansorahan/spatialite"
	// _ "github.com/mattn/go-sqlite3"
	// _ "github.com/shaxbee/go-spatialite"
	// "github.com/shaxbee/go-spatialite/wkb"
	// "github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
)

func makeDB() *sql.DB {
	// db, err := sql.Open("sqlite3", "c:/temp/China.udd")
	db, err := sql.Open("spatialite", "c:/temp/xian.udbx")
	// require.NoError(t, err)
	fmt.Println("db: ", db, err)

	res, err := db.Exec("SELECT InitSpatialMetadata()")
	fmt.Println("res: ", res, err)
	// require.NoError(t, err)
	return db
}

func TestSpatialite() {
	db := makeDB()
	defer db.Close()

	res, err := db.Exec("CREATE TABLE poi(title TEXT)")
	fmt.Println("1", res, err)
	// require.NoError(t, err)

	res, err = db.Exec("SELECT AddGeometryColumn('poi', 'loc', 4326, 'POINT')")
	fmt.Println("2", res, err)
	// require.NoError(t, err)

	// p1 := wkb.Point{10, 10}
	// res, err = db.Exec("INSERT INTO poi(title, loc) VALUES (?, ST_PointFromWKB(?, 4326))", "foo", p1)
	// fmt.Println("3", res, err)
	// // assert.NoError(t, err)

	// p2 := wkb.Point{}
	// r := db.QueryRow("SELECT ST_AsBinary(loc) AS loc FROM poi WHERE title=?", "foo")
	// err = r.Scan(&p2)
	// fmt.Println("4", r, err)

	// if assert.NoError(t, err) {
	// assert.Equal(t, p1, p2)
	// }
}

func test2() {
	db, err := sql.Open("spatialite", "c:/temp/xian.udbx")
	// require.NoError(t, err)
	fmt.Println("db: ", db, err)

	rows, err := db.Query("SELECT name FROM sqlite_master ")
	fmt.Println("rows: ", rows, err)
	// defer rows.Close()

	// res, err := db.Exec(`CREATE TABLE IF NOT EXISTS spatialite_test (test_geom ST_Geometry)`)
	// fmt.Println("res: ", res, err)
}

func main() {
	test2()
	// TestSpatialite()
	// testSqlite()
}

func testSqlite() {
	db, err := sql.Open("sqlite3", "c:/temp/China.udd")
	defer db.Close()

	fmt.Println("db: ", db, err)
	if err = db.Ping(); err != nil {
		fmt.Println("db error: ", err)
	}
	rows, err := db.Query("SELECT name FROM sqlite_master ")
	defer rows.Close()

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			fmt.Println(err) // Handle scan error
		}
		fmt.Println("name: ", name)
		r2, _ := db.Query("select * from " + name)
		cs, _ := r2.ColumnTypes()
		for _, c := range cs {
			fmt.Println("columns: ", c.Name())
		}
	}

	// db.
}
