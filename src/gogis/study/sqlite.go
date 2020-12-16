package main

import (
	"database/sql"
	"fmt"
	"gogis/base"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	// "github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
)

func test211() {
	db, err := sql.Open("spatialite", "c:/temp/xian.udbx")
	// require.NoError(t, err)
	fmt.Println("db: ", db, err)

	rows, err := db.Query("SELECT name FROM sqlite_master ")
	fmt.Println("rows: ", rows, err)
	// defer rows.Close()

	// res, err := db.Exec(`CREATE TABLE IF NOT EXISTS spatialite_test (test_geom ST_Geometry)`)
	// fmt.Println("res: ", res, err)
}

func sqlmain() {
	// test2()
	// TestSpatialite()
	testUDBX()
	// testCreate()
}

func testCreate() {
	db, err := sql.Open("sqlite3", "c:/temp/temp.db")
	defer db.Close()

	fmt.Println("db: ", db, err)
	if err = db.Ping(); err != nil {
		fmt.Println("db error: ", err)
	}

	sql_table := `
    CREATE TABLE IF NOT EXISTS userinfo(
        uid INTEGER PRIMARY KEY AUTOINCREMENT,
        username VARCHAR(64) NULL,
        departname VARCHAR(64) NULL,
        created DATE NULL
    );
    `

	res, err := db.Exec(sql_table)
	fmt.Println("Exec res:", res, "error: ", err)

}

func testUDBX() {
	db, err := sql.Open("sqlite3", "c:/temp/temp.udbx")
	defer db.Close()

	fmt.Println("db: ", db, err)
	if err = db.Ping(); err != nil {
		fmt.Println("db error: ", err)
	}
	table := "JBNTBHTB"
	var objCount int64
	db.QueryRow("select count(*) from " + table).Scan(&objCount)
	oneTable(db, table, 0, int(objCount), nil)
	return

	oneCount := 200000
	conCount := (int(objCount) / oneCount) + 1
	tr := base.NewTimeRecorder()

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < conCount; i++ {
		wg.Add(1)
		go oneTable(db, table, i*oneCount, oneCount*(i+1), wg)
	}
	wg.Wait()

	tr.Output("scan table" + table)
}

func oneTable(db *sql.DB, table string, min, max int, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	tr := base.NewTimeRecorder()
	// r2, err := db.Query("select SmID,SmGeometry from " + table + " where SmID>=" + strconv.Itoa(min) + " and SmID <" + strconv.Itoa(max))
	r2, err := db.Query("select * from " + table)
	defer r2.Close()
	tr.Output("just query")

	if err == nil {
		// cs, err := r2.ColumnTypes()
		// if err == nil {
		// 	for _, c := range cs {
		// 		fmt.Println("columns: ", c.Name())
		// 	}
		// } else {
		// 	fmt.Println("column types error:", err)
		// }

		show := false

		var id int
		var data []byte
		for r2.Next() {
			err := r2.Scan(&id, &data)
			if err == nil {
				if !show {
					fmt.Println("smid:", id)
					// fmt.Println("smid:", id, "smgeometry:", data)
					show = true
				}
			} else {
				// fmt.Println("Scan error:", err)
				// return
			}
		}
	} else {
		fmt.Println("select error:", err) // Handle scan error
	}

}

func testSqlite() {
	db, err := sql.Open("sqlite3", "c:/temp/temp.udbx")
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
			fmt.Println("scan error:", err) // Handle scan error
		}
		fmt.Println("name: ", name)
		// oneTable(db, name)
	}

	// db.
}
