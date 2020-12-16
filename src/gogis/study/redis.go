package main

// import (
// 	"fmt"
// 	"gogis/base"
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"

// 	// "github.com/syndtr/goleveldb/leveldb"
// 	"context"

// 	"github.com/go-redis/redis"
// )

// var gPath = "c:/temp/"
// var gTitle = "JBNTBHTB"
// var ctx = context.Background()

// func main() {
// 	testWriteTile()
// 	// testReadTile()

// 	return

// }

// var dbPath = gPath + gTitle + ".db"

// var db *redis.Client

// func testWriteTile() {
// 	var opt redis.Options
// 	opt.Addr = "localhost:6379"
// 	db = redis.NewClient(&opt)
// 	_, err := db.Ping(ctx).Result()
// 	if err != nil {
// 		fmt.Println("open redis:", err)
// 	}

// 	tr := base.NewTimeRecorder()
// 	filepath.Walk("c:/temp/cache/JBNTBHTB/", WalkFn)
// 	tr.Output("write leveldb")
// 	db.Close()
// }

// func WalkFn(path string, f os.FileInfo, err error) error {
// 	if err != nil {
// 		return err
// 	}
// 	if f == nil || f.IsDir() {
// 		return nil
// 	}

// 	data, _ := ioutil.ReadFile(path)

// 	db.Set(ctx, path, data, 100)

// 	return nil
// }

// // func WalkFnRead(path string, f os.FileInfo, err error) error {
// // 	if err != nil {
// // 		return err
// // 	}
// // 	if f == nil || f.IsDir() {
// // 		return nil
// // 	}

// // 	db.Get([]byte(path), nil)
// // 	// fmt.Println(data)

// // 	// ioutil.ReadFile(path)
// // 	// db.Put([]byte(path), buf, nil)

// // 	return nil
// // }

// // func testReadTile() {
// // 	db, _ = leveldb.OpenFile(dbPath, nil)
// // 	tr := base.NewTimeRecorder()
// // 	filepath.Walk("c:/temp/cache/JBNTBHTB/", WalkFnRead)
// // 	// iter := db.NewIterator(nil, nil)
// // 	// // iter.
// // 	// for iter.Next() {
// // 	// 	// Remember that the contents of the returned slice should not be modified, and
// // 	// 	// only valid until the next call to Next.
// // 	// 	key := iter.Key()
// // 	// 	fmt.Println(string(key))
// // 	// 	// base.Bytes2Int(key)
// // 	// 	iter.Value()
// // 	// }
// // 	tr.Output("read leveldb")
// // }
