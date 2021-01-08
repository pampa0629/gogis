package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"gogis/base"
	"io"
	"strconv"
	"sync"
	"time"

	pool "github.com/silenceper/poor"
	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"
)

var hbasePool pool.Pool

func initPoor(address string) {
	// factory := func() (interface{}, error) { return gohbase.NewClient("localhost:2181"), nil }
	factory := func() (interface{}, error) { return gohbase.NewClient(address), nil }

	//close 关闭链接的方法
	close := func(v interface{}) error { v.(gohbase.Client).Close(); return nil }

	//创建一个连接池
	poolConfig := &pool.Config{
		InitialCap: 5,
		MaxIdle:    200,
		MaxCap:     200,
		Factory:    factory,
		Close:      close,
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: 15 * time.Second,
	}
	hbasePool, _ = pool.NewChannelPool(poolConfig)
	// if err != nil {
	// 	fmt.Println("err=", err)
	// }
}

func hbase_main() {
	scankey3()
	// createTable()
	// testCount()
	// writeTable()
	// listTables()
}

func scankey() {
	initPoor("localhost:2181")
	tr := base.NewTimeRecorder()

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go batchscan(i, wg)
	}
	wg.Wait()

	tr.Output("scan all")
}

func batchscan(num int, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	tr := base.NewTimeRecorder()

	// client := gohbase.NewClient("localhost:2181")
	v, err := hbasePool.Get()
	if v == nil || err != nil {
		fmt.Println("pool get error:", err)
	}
	client, _ := v.(gohbase.Client)

	start := getRowkey(0, 0)
	end := getRowkey(1, 0)
	// fmt.Println("start:", start, "end:", end)

	scanRequest, err := hrpc.NewScanRange(context.Background(), []byte("DLTB"), start, end)
	if err != nil {
		fmt.Println("NewScanRange error:", err)
	}
	scan := client.Scan(scanRequest)
	count := 0
	for {
		getRsp, err := scan.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		if err != nil {
			fmt.Println("scan next error:", err)
		}
		fmt.Println("rowkey:", getRsp.Cells[0].Row)
		// for i, v := range getRsp.Cells {
		// 	fmt.Println("i:", i, "rowkey:", v.Row)
		// }
		count++
	}
	if count != 1307 {
		fmt.Println("count:", count, "start:", start, "end:", end)
	}
	// client.Close()
	hbasePool.Put(client)
	tr.Output("scan one: " + strconv.Itoa(num))
}

func getRowkey(code int32, id int64) []byte {

	// io.Closer

	var buf = make([]byte, 12)
	// 这里必须用big，保证 code+1时，是在末端 增加 byte，而不是前端，保证row key的有序
	binary.BigEndian.PutUint32(buf, uint32(code))
	binary.BigEndian.PutUint64(buf[4:], uint64(id))
	return buf
}

func scankey3() {
	client := gohbase.NewClient("localhost:2181")
	start := getRowkey(0, 0)
	end := getRowkey(1365, 0)
	fmt.Println("start:", start, "end:", end)

	// scanRequest, _ := hrpc.NewScanRange(context.Background(), []byte("JBNTBHTB"), start, end)
	scanRequest, _ := hrpc.NewScan(context.Background(), []byte("JBNTBHTB"))
	scan := client.Scan(scanRequest)
	count := 0
	for {
		getRsp, err := scan.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		if err != nil {
			fmt.Println("scan next error:", err)
		}
		// for i, v := range getRsp.Cells {
		// 	fmt.Println("i:", i, "rowkey:", v.Row)
		// }

		count++
	}
	fmt.Println("count:", count, "start:", start, "end:", end)
	client.Close()
}

func testCount() {
	tr := base.NewTimeRecorder()
	client := gohbase.NewClient("localhost:2181")
	start := "00a64321-7fa7-469a-bb9c-f5dd8f67f5ed"
	end := "b0a64328-5bb5-4742-ab49-6701ad4c09ff"
	// scanRequest, err := hrpc.NewScanRange(context.Background(), []byte("temp_JBNTBHTB_id"), []byte{0}, []byte{255})
	scanRequest, err := hrpc.NewScanRangeStr(context.Background(), "temp_JBNTBHTB_id", start, end)
	scan := client.Scan(scanRequest)
	tr.Output("scan")
	if err == nil {
		count := 0
		for {
			getRsp, err := scan.Next()
			if err == io.EOF || getRsp == nil {
				break
			}
			if err != nil {
				fmt.Println("scan next error:", err)
			}
			count++
		}
		fmt.Println("scan count :", count)
	} else {
		fmt.Println("client get error:", err)
	}
	client.Close()
	tr.Output("count")
}

func testInc() {
	tr := base.NewTimeRecorder()

	client := gohbase.NewClient("localhost:2181")

	tr.Output("connect")
	mutate, err := hrpc.NewIncSingle(context.Background(), []byte("chinapnt_84"), []byte{100}, "geo", "geom", 100)
	if err != nil {
		fmt.Println("NewScanStr error:", err)
	}
	// scanRequest, err := hrpc.NewScanStr(context.Background(), "temp_chinapnt_5f84_id")
	count, err := client.Increment(mutate)

	if err != nil {
		fmt.Println("NewScanStr error:", err)
	}
	fmt.Println("count:", count)
	client.Close()
	tr.Output("Increment")
}

func listTables() {
	tr := base.NewTimeRecorder()
	host := "localhost:2181"
	adminClient := gohbase.NewAdminClient(host)
	tr.Output("admin")
	return
	listTNs, _ := hrpc.NewListTableNames(context.Background())
	tableNames, err := adminClient.ListTableNames(listTNs)
	fmt.Println(err)
	for _, v := range tableNames {
		fmt.Println(v)
	}

}

func createTable() {
	host := "localhost:2181"
	adminClient := gohbase.NewAdminClient(host)
	// adminClient.

	tableName := "test2" // 要创建的表名
	families := map[string]map[string]string{"cf": {}}
	ct := hrpc.NewCreateTable(context.Background(), []byte(tableName), families)
	err := adminClient.CreateTable(ct)
	fmt.Println("create table error:", err)
}

func writeTable2() {
	tr := base.NewTimeRecorder()
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	start := 60000
	count := 2000
	num := 1000
	for i := start; i < start+count; i++ {
		wg.Add(1)
		go writeTableBatch(i*num, num, wg)
	}
	wg.Wait()
	tr.Output("write go")
}

func writeTable() {

	host := "localhost:2181"
	client := gohbase.NewClient(host)

	// rowKey := 0 // RowKey
	tr := base.NewTimeRecorder()
	for i := 0; i < 100; i++ {
		// var buf = make([]byte, 4)
		// binary.LittleEndian.PutUint32(buf, uint32(i))

		var buf = make([]byte, 12)
		// 这里必须用big，保证 code+1时，是在末端 增加 byte，而不是前端，保证row key的有序
		binary.BigEndian.PutUint32(buf, uint32(i))
		binary.BigEndian.PutUint64(buf[4:], uint64(0))

		value := map[string]map[string][]byte{
			"cf": { // 列族名, 与创建表时指定的名字保持一致
				"col1": buf,
			},
		}

		putRequest, err := hrpc.NewPut(context.Background(), []byte("test2"), buf, value)
		// putRequest.SkipBatch()
		if err != nil {
			fmt.Println("new put error:", err)
		}
		// fmt.Println("write error:", err)
		_, err = client.Put(putRequest)
		if err != nil {
			fmt.Println("put error:", err)
		}
	}

	tr.Output("write hbase")
	client.Close()
	// return
	// fmt.Println("put res:", res)
}

func writeTableBatch(start int, count int, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	host := "localhost:2181"
	client := gohbase.NewClient(host)

	// rowKey := 0 // RowKey
	value := map[string]map[string][]byte{
		"cf": { // 列族名, 与创建表时指定的名字保持一致
			"col1": []byte("val1"), // 列与值, 列名可自由定义
			// "col2": []byte("val2"),
			// "col3": []byte("val3"),
		},
	}
	tr := base.NewTimeRecorder()
	for i := start; i < start+count; i++ {
		putRequest, err := hrpc.NewPutStr(context.Background(), "test1", strconv.Itoa(i), value)
		// putRequest.SkipBatch()
		if err != nil {
			fmt.Println("new put error:", err)
		}
		// fmt.Println("write error:", err)
		_, err = client.Put(putRequest)
		if err != nil {
			fmt.Println("put error:", err)
		}
	}

	tr.Output("write hbase")
	client.Close()
	// return
	// fmt.Println("put res:", res)
}

func scankey2() {
	tr := base.NewTimeRecorder()

	client := gohbase.NewClient("localhost:2181")

	tr.Output("connect")
	scanRequest, err := hrpc.NewScanStr(context.Background(), "test1")

	if err != nil {
		fmt.Println("NewScanStr error:", err)
	}
	scan := client.Scan(scanRequest)

	// var res []*hrpc.Result
	count := 0
	for {
		getRsp, err := scan.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		if err != nil {
			// log.Print(err, "scan.Next")
			fmt.Println("scan next error:", err)
		}
		// fmt.Println("cells:", getRsp.Cells)
		for i, v := range getRsp.Cells {
			fmt.Println("i:", i)
			fmt.Println("row:", v.Row)
			// fmt.Println("Family:", v.Family)
			// fmt.Println("CellType:", v.CellType)
			// fmt.Println("Qualifier:", v.Qualifier)
			// fmt.Println("Tags:", v.Value)
			fmt.Println("Value:", v.Value)
		}
		count++
	}
	fmt.Println("scan count:", count)
	tr.Output("get all")
}

// func scankey() {
// 	// client := gohbase.NewClient("localhost:2181")
// 	tr := base.NewTimeRecorder()

// 	keys := []string{
// 		"00a64321-7fa7-469a-bb9c-f5dd8f67f5ed",
// 		"00a64328-5b2d-4820-a6b5-fa75736e8c39",
// 		"00a64328-5baf-4880-9d67-6faedddec677",
// 		"10a64321-7fb5-4252-a11f-01d858c836ce",
// 		"10a64328-5b36-4073-884b-88afe4ac67b9",
// 		"10a64328-5bb6-4550-acbf-7faa3309538a",
// 		"20a64321-7fbc-4088-a7b4-4be0b45e64d6",
// 		"20a64328-5b3d-4efe-9dc1-3f88c60ad5f5",
// 		"20a64328-5bbd-4e57-9a3b-97b3f9249caa",
// 		"30a64321-7fbd-4c3a-9365-d02c7fd55350",
// 		"30a64328-5b60-4beb-8b41-b4bfdc0af752",
// 		"30a64328-5be5-416c-9143-498e751f2515",
// 		"40a64321-7fbf-4a22-94c0-d21fed0d3349",
// 		"40a64328-5b63-495d-80d8-6d4671cdf995",
// 		"40a64328-5bec-4164-a919-ed996a2a683e",
// 		"50a64321-7ff0-4d1d-a549-7f66f584f827",
// 		"50a64328-5b67-436d-a0a7-de4d8bc75849",
// 		"50a64328-5bf1-4da9-9fa4-8820b7fac190",
// 		"60a64321-7ff2-4fc6-bbf0-edb759e87526",
// 		"60a64328-5b74-408e-a33f-84391f40f2f2",
// 		"60a64328-5bfc-48d1-8f52-1006bd1e256b",
// 		"70a64328-597c-4637-b404-eac00901c76d",
// 		"70a64328-5b75-47aa-a513-271924f57691",
// 		"70a64328-5f2d-4ec7-9ef3-89901f585212",
// 		"80a64328-59ec-44b5-93b6-51ca1b0d5dc2",
// 		"80a64328-5b7c-4171-afe5-9da218044b90",
// 		"80a64328-5f61-4fec-a864-f35318ef9008",
// 		"90a64328-59f5-467f-8422-2c43e70759d2",
// 		"90a64328-5ba7-4191-83b1-adca4f260778",
// 		"90a64328-5f75-4059-9200-ac0591503db2",
// 		"a0a64328-5b26-41ab-9ae5-c24f84313175",
// 		"a0a64328-5bae-41dd-b268-326e652df753",
// 		"b0a64321-7fb4-431c-a470-c3416b9be065",
// 		"b0a64328-5b34-430f-bb35-048be1ca387f",
// 		"b0a64328-5bb5-4742-ab49-6701ad4c09ff",
// 		"c0a64321-7fb7-4554-bd88-4f0fab391242",
// 		"c0a64328-5b3d-428c-a46b-92e075b4bb35",
// 		"c0a64328-5bbd-4848-8caf-b96ddf14aa82",
// 		"d0a64321-7fbd-4815-9479-30900b8a0aaa",
// 		"d0a64328-5b60-4b04-9da3-7516053c40a6",
// 		"d0a64328-5be4-4875-847d-a3bcc2611156",
// 		"e0a64321-7fbe-4e79-a6e4-bed5daa1c723",
// 		"e0a64328-5b62-49a9-9196-c1c8369bcc62",
// 		"e0a64328-5be6-4cd3-aea9-b4b417a12563",
// 		"f0a64321-7fe2-4df2-8845-8de99b4f87c7",
// 		"f0a64328-5b66-448f-809f-3bf03fc6e8ff",
// 		"f0a64328-5bef-4127-bc95-b28ad811c6e7",
// 		"f0a64328-5f75-40cb-ac05-7e1da612b775"}

// 	var wg *sync.WaitGroup = new(sync.WaitGroup)
// 	skip := 10
// 	for i := 0; i < len(keys); i += skip {
// 		wg.Add(1)
// 		end := i + skip
// 		if end > len(keys) {
// 			end = len(keys) - 1
// 		}
// 		// go scankeyrange(client, "temp_JBNTBHTB_id", string(keys[i][0]), string(keys[end][0]), wg)
// 	}
// 	wg.Wait()

// 	tr.Output("scan all")

// }
