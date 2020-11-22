package main

import (
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/geometry"
	"sync"
	"time"

	// "github.com/gomodule/redigo/redis"
	// "github.com/go-redis/redis"
	"github.com/syndtr/goleveldb/leveldb"
)

func SaveDB() {
	var gPath = "c:/temp/"
	var gTitle = "JBNTBHTB"
	var gExt = ".shp"

	dbPath := gPath + gTitle + ".db"
	db, _ := leveldb.OpenFile(dbPath, nil)

	feset := data.OpenShape(gPath + gTitle + gExt)
	feait := feset.QueryByBounds(feset.GetBounds())
	fmt.Println("count:", feait.Count())

	startTime := time.Now().UnixNano()

	id := int32(0)
	for {
		fea, ok := feait.Next()
		if ok {
			data := fea.Geo.To(geometry.WKB)
			db.Put(base.Int2Bytes(id), data, nil)
			// fmt.Println("id:", id, data)
			// fmt.Println("geo:", fea.Geo)
			id++
		} else {
			fmt.Println("id:", id)
			break
		}
	}
	db.Close()

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("time: %f 毫秒", seconds)
}

const gBatchCount = 10000

func batchOpenDB(id int32, db *leveldb.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := id; i < id+gBatchCount; i++ {
		data, err := db.Get(base.Int2Bytes(i), nil)
		if err != nil {
			fmt.Println("get error:", err)
		}
		var geo geometry.GeoPolygon
		geo.From(data, geometry.WKB)
		// fmt.Println("id:", i)
		// fmt.Println(geoPoint)
	}
}

func OpenDB() {
	startTime := time.Now().UnixNano()

	var gPath = "c:/temp/"
	var gTitle = "JBNTBHTB"

	dbPath := gPath + gTitle + ".db"
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		fmt.Println("open error:", err)
	}

	count := 2320000
	batch := int(count / gBatchCount)

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < batch; i++ {
		wg.Add(1)
		go batchOpenDB((int32)(i*gBatchCount), db, wg)
	}
	wg.Wait()

	// for i := int32(0); i < 209355; i++ {
	// 	data, err := db.Get(base.Int2Bytes(i), nil)
	// 	if err != nil {
	// 		fmt.Println("get error:", err)
	// 	}
	// 	var geo geometry.GeoPolygon
	// 	geo.From(data, geometry.WKB)
	// 	// fmt.Println("id:", i)
	// 	// fmt.Println(geoPoint)
	// }

	// iter := db.NewIterator(nil, nil)
	// // iter.
	// id := 0
	// for iter.Next() {
	// 	// Remember that the contents of the returned slice should not be modified, and
	// 	// only valid until the next call to Next.
	// 	key := iter.Key()
	// 	base.Bytes2Int(key)
	// 	value := iter.Value()
	// 	var geo geometry.GeoPolygon
	// 	geo.From(value, geometry.WKB)

	// 	id++
	// 	// if id%10000 == 0 {
	// 	// 	fmt.Println(geo.GetBounds())
	// 	// }
	// 	// fmt.Println("id:", id)

	// 	// ...
	// }
	// fmt.Println("id:", id)
	// iter.Release()
	// err = iter.Error()
	// fmt.Println("error:", err)

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("time: %f 毫秒", seconds)

}

func OpenShapeMem() {
	startTime := time.Now().UnixNano()

	var gPath = "c:/temp/"
	var gTitle = "JBNTBHTB"
	var gExt = ".shp"

	feset := data.OpenShape(gPath + gTitle + gExt)
	feait := feset.QueryByBounds(feset.GetBounds())
	for {
		fea, ok := feait.Next()
		if ok {
			geo := fea.Geo
			geo.GetBounds()
		} else {
			break
		}
	}

	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e6)
	fmt.Printf("time: %f 毫秒", seconds)
}

func level_main() {
	// SaveDB()
	OpenDB()
	// OpenShapeMem()

	return

}
