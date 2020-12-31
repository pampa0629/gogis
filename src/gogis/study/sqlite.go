package main

import (
	"database/sql"
	"fmt"
	"gogis/base"
	"strconv"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	// "github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
)

func sqlmain() {
	// test2()
	testSpatialite2()
	// testUDBX()
	// testCreate()
}

func testSpatialite2() {
	tr := base.NewTimeRecorder()
	filename := "c:/temp/DLTB.sqlite"
	db, _ := sql.Open("sqlite3", filename)
	defer db.Close()
	tr.Output("open db " + filename)

	// table := "JBNTBHTB"
	table := "DLTB" // SpatialIndex chinapnt_84
	var objCount int64
	db.QueryRow("select count(*) from " + table).Scan(&objCount)
	fmt.Println("obj count:", objCount)

	// oneSpatialiteTable2(db, table, 0, int(objCount), nil)

	// tr.Output("scan table one time")
	// return

	oneCount := 10000
	conCount := (int(objCount) / oneCount) + 1

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < conCount; i++ {
		wg.Add(1)
		go oneSpatialiteTable3(filename, table, i*oneCount, oneCount*(i+1), wg)
	}
	wg.Wait()

	tr.Output("go scan table " + table)
}

func oneSpatialiteTable3(filename, table string, min, max int, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	// filename := "c:/temp/DLTB.sqlite"
	db, _ := sql.Open("sqlite3", filename)
	defer db.Close()

	where := " where rowid between " + strconv.Itoa(min) + " and " + strconv.Itoa(max)

	sql := "select rowid,geom from " + table + where
	fmt.Println("sql:", sql)

	tr := base.NewTimeRecorder()
	r2, err := db.Query(sql)
	fmt.Println("db query error:", err)
	// r2, err := db.Query("select * from " + table)
	defer r2.Close()
	tr.Output("just query")

	if err == nil {
		var id int
		var data []byte
		for r2.Next() {
			err := r2.Scan(&id, &data)
			if err != nil {
				fmt.Println("Scan error:", err)
			}
		}
	} else {
		fmt.Println("select error:", err) // Handle scan error
	}
	tr.Output("scan table")
}

func oneSpatialiteTable2(db *sql.DB, table string, min, max int, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	where := " where rowid between " + strconv.Itoa(min) + " and " + strconv.Itoa(max)

	sql := "select rowid,geom from " + table + where
	fmt.Println("sql:", sql)

	tr := base.NewTimeRecorder()
	r2, err := db.Query(sql)
	fmt.Println("db query error:", err)
	// r2, err := db.Query("select * from " + table)
	defer r2.Close()
	tr.Output("just query")

	if err == nil {
		var id int
		var data []byte
		for r2.Next() {
			err := r2.Scan(&id, &data)
			if err != nil {
				fmt.Println("Scan error:", err)
			}
		}
	} else {
		fmt.Println("select error:", err) // Handle scan error
	}
	tr.Output("scan table")
}

// =========================================

func testSpatialite() {
	db, err := sql.Open("sqlite3", "c:/temp/chinapnt_84.sqlite")
	defer db.Close()

	fmt.Println("db: ", db, err)
	if err = db.Ping(); err != nil {
		fmt.Println("db error: ", err)
	}
	// table := "JBNTBHTB"
	table := "vector_layers_statistics" // SpatialIndex chinapnt_84
	var objCount int64
	db.QueryRow("select count(*) from " + table).Scan(&objCount)
	fmt.Println("obj count:", objCount)
	oneSpatialiteTable(db, table, 0, int(objCount), nil)
	return

	oneCount := 1000000
	conCount := (int(objCount) / oneCount) + 1
	tr := base.NewTimeRecorder()

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < conCount; i++ {
		wg.Add(1)
		go oneTable(db, table, i*oneCount, oneCount*(i+1), wg)
	}
	wg.Wait()

	tr.Output("scan table " + table)
}

func oneSpatialiteTable(db *sql.DB, table string, min, max int, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	// inter := (max - min) / 20
	// where := " where smid in " + in
	// count := 2

	// for i := 0; i < count; i++ {
	// 	where += "select smid,smgeometry from " + table + " where SmID BETWEEN " + strconv.Itoa(min) + " AND " + strconv.Itoa(min+inter) + " "
	// 	min += 2 * inter
	// 	if i != count-1 {
	// 		where += " union " // 还不如用 or
	// 		// 回头再试试 in OK
	// 	}
	// }
	sql := "select row_count,extent_min_x,extent_min_y,extent_max_x,extent_max_y from " + table //  + where
	// sql := where
	fmt.Println("sql:", sql)

	// var objCount int64
	// db.QueryRow("select count(*) from " + table + where).Scan(&objCount)
	// fmt.Println("obj count:", objCount)
	// tr.Output("count")
	tr := base.NewTimeRecorder()
	r2, err := db.Query(sql)
	fmt.Println("db query error:", err)
	// r2, err := db.Query("select * from " + table)
	defer r2.Close()
	tr.Output("just query")

	if err == nil {
		cs, err := r2.ColumnTypes()
		if err == nil {
			for _, c := range cs {
				fmt.Println("columns: ", c.Name(), c.DatabaseTypeName())
			}
		} else {
			fmt.Println("column types error:", err)
		}

		// show := false

		// var id int
		var count int
		var extents [4]float64
		// var data []byte
		for r2.Next() {
			err := r2.Scan(&count, &extents[0], &extents[1], &extents[2], &extents[3])
			if err == nil {
				// if !show {
				// fmt.Println("smid:", id)
				// if id%1000 == 0 {
				// 	fmt.Println("smid:", id)
				// }
				fmt.Println("count:", count, "extents:", extents)
				// show = true
				// }
			} else {
				fmt.Println("Scan error:", err)
				// return
			}
		}
	} else {
		fmt.Println("select error:", err) // Handle scan error
	}
	tr.Output("scan table")
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
	db, err := sql.Open("sqlite3", "c:/temp/chinapnt_84.sqlite")
	defer db.Close()

	fmt.Println("db: ", db, err)
	if err = db.Ping(); err != nil {
		fmt.Println("db error: ", err)
	}
	// table := "JBNTBHTB"
	table := "chinapnt_84"
	var objCount int64
	db.QueryRow("select count(*) from " + table).Scan(&objCount)
	fmt.Println("obj count:", objCount)
	oneTable(db, table, 0, int(objCount), nil)
	return

	oneCount := 1000000
	conCount := (int(objCount) / oneCount) + 1
	tr := base.NewTimeRecorder()

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < conCount; i++ {
		wg.Add(1)
		go oneTable(db, table, i*oneCount, oneCount*(i+1), wg)
	}
	wg.Wait()

	tr.Output("scan table " + table)
}

func oneTable(db *sql.DB, table string, min, max int, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	in := "("
	count := 10
	for i := 0; i < count; i++ {
		in += strconv.Itoa(min + i)
		if i != count-1 {
			in += ","
		}
	}
	in += ")"

	// inter := (max - min) / 20
	where := " where smid in " + in
	// count := 2

	// for i := 0; i < count; i++ {
	// 	where += "select smid,smgeometry from " + table + " where SmID BETWEEN " + strconv.Itoa(min) + " AND " + strconv.Itoa(min+inter) + " "
	// 	min += 2 * inter
	// 	if i != count-1 {
	// 		where += " union " // 还不如用 or
	// 		// 回头再试试 in OK
	// 	}
	// }
	sql := "select SmID, SmGeometry from " + table + where
	// sql := where
	fmt.Println("sql:", sql)

	// var objCount int64
	// db.QueryRow("select count(*) from " + table + where).Scan(&objCount)
	// fmt.Println("obj count:", objCount)
	// tr.Output("count")
	tr := base.NewTimeRecorder()
	r2, err := db.Query(sql)
	fmt.Println("db query error:", err)
	// r2, err := db.Query("select * from " + table)
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

		// show := false

		var id int
		var data []byte
		for r2.Next() {
			err := r2.Scan(&id, &data)
			if err == nil {
				// if !show {
				// fmt.Println("smid:", id)
				// if id%1000 == 0 {
				// 	fmt.Println("smid:", id)
				// }
				fmt.Println("smid:", id, "smgeometry:", data)
				// show = true
				// }
			} else {
				// fmt.Println("Scan error:", err)
				// return
			}
		}
	} else {
		fmt.Println("select error:", err) // Handle scan error
	}
	tr.Output("scan table")
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
