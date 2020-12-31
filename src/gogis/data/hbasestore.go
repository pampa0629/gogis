package data

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"gogis/index"
	"io"
	"sync"
	"time"

	pool "github.com/silenceper/poor"
	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"
)

type HbaseStore struct {
	address string // zookeeper's address, such as: localhost:2182 or ip:2181
	feasets []*HbaseFeaset
	hpool   pool.Pool
}

const HBASE_SYS_TABLE = "gogis_sys"

func (this *HbaseStore) initPoor() {
	factory := func() (interface{}, error) { return gohbase.NewClient(this.address), nil }
	//close 关闭链接的方法
	close := func(v interface{}) error { v.(gohbase.Client).Close(); return nil }

	//创建一个连接池
	poolConfig := &pool.Config{
		InitialCap: 10,
		MaxIdle:    100,
		MaxCap:     100,
		Factory:    factory,
		Close:      close,
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: 15 * time.Second,
	}
	this.hpool, _ = pool.NewChannelPool(poolConfig)
}

func (this *HbaseStore) Open(params ConnParams) (bool, error) {
	this.address = params["address"]
	if this.address == "" {
		return false, errors.New("must set 'address' of ConnParams as 'ip:2181'.")
	}
	this.initPoor() // 初始化数据库连接池

	// 判断是否有 gogis_sys 表
	adminClient := gohbase.NewAdminClient(this.address)
	if this.hasSysTable(adminClient) {
		// 读取系统表，创建 [] feaset
		this.loadSysTable()
	} else {
		// 创建系统表
		this.createSysTable(adminClient)
	}

	return true, nil
}

func (this *HbaseStore) openClient() gohbase.Client {
	v, _ := this.hpool.Get()
	client, _ := v.(gohbase.Client)
	return client
}

func (this *HbaseStore) closeClient(client gohbase.Client) {
	this.hpool.Put(client)
}

func (this *HbaseStore) loadSysTable() {
	// client := gohbase.NewClient(this.address)
	client := this.openClient()

	scanRequest, _ := hrpc.NewScanStr(context.Background(), HBASE_SYS_TABLE)
	scan := client.Scan(scanRequest)

	this.feasets = make([]*HbaseFeaset, 0)
	for {
		getRsp, err := scan.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		feaset := new(HbaseFeaset)
		feaset.store = this
		for _, v := range getRsp.Cells {
			feaset.name = string(v.Row)
			// "geo": { // 列族名, 与创建表时指定的名字保持一致
			// 	"bbox":        this.bbox.ToBytes(),
			// 	"geotype":     base.Int32ToBytes(int32(this.geotype)),
			// 	"index_level": base.Int32ToBytes(HBASE_INDEX_LEVEL),
			// 	"count":       base.Int32ToBytes(int32(this.count)),
			// },
			if string(v.Family) == "geo" {
				if string(v.Qualifier) == "bbox" {
					feaset.bbox.FromBytes(v.Value)
				} else if string(v.Qualifier) == "geotype" {
					feaset.geotype = geometry.GeoType(base.BytesToInt32(v.Value))
				} else if string(v.Qualifier) == "index_level" {
					feaset.index_level = base.BytesToInt32(v.Value)
				} else if string(v.Qualifier) == "count" {
					feaset.count = base.BytesToInt64(v.Value)
				}
			}
		}
		this.feasets = append(this.feasets, feaset)
	}
	this.closeClient(client)
}

func (this *HbaseStore) createSysTable(adminClient gohbase.AdminClient) {
	families := map[string]map[string]string{"geo": {}}
	ct := hrpc.NewCreateTable(context.Background(), []byte(HBASE_SYS_TABLE), families)
	adminClient.CreateTable(ct)
}

// 判断是否有系统表
func (this *HbaseStore) hasSysTable(adminClient gohbase.AdminClient) bool {
	listTNs, _ := hrpc.NewListTableNames(context.Background())
	tables, _ := adminClient.ListTableNames(listTNs)
	for _, v := range tables {
		if string(v.Qualifier) == HBASE_SYS_TABLE {
			return true
		}
	}
	return false
}

// 创建要素表
func (this *HbaseStore) createFeaTable(name string) bool {
	adminClient := gohbase.NewAdminClient(this.address)
	families := map[string]map[string]string{"geo": {}}
	ct := hrpc.NewCreateTable(context.Background(), []byte(name), families)
	adminClient.CreateTable(ct)
	return true
}

// 删除要素表
func (this *HbaseStore) deleteFeaTable(name string) {
	adminClient := gohbase.NewAdminClient(this.address)
	dis := hrpc.NewDisableTable(context.Background(), []byte(name))
	adminClient.DisableTable(dis)
	del := hrpc.NewDeleteTable(context.Background(), []byte(name))
	adminClient.DeleteTable(del)
}

// 删除系统表中的相关记录
func (this *HbaseStore) deleteSysInfo(name string) {
	client := this.openClient()
	req, _ := hrpc.NewDelStr(context.Background(), HBASE_SYS_TABLE, name, nil)
	client.Delete(req)
	this.closeClient(client)
}

// 删除要素集
func (this *HbaseStore) DeleteFeaset(name string) (bool, error) {
	feaset, _ := this.GetFeasetByName(name)
	if feaset != nil {
		// 删除 要素表
		this.deleteFeaTable(name)

		// 删除系统表中的相关记录
		this.deleteSysInfo(name)

		// 从[]中删除
		for i, v := range this.feasets {
			if v == feaset {
				this.feasets = append(this.feasets[:i], this.feasets[i+1:]...)
				break
			}
		}
		return true, nil
	}
	return false, errors.New("feaset name: " + name + " is not existed.")
}

// 创建要素集
func (this *HbaseStore) CreateFeaset(name string, bbox base.Rect2D, geotype geometry.GeoType) Featureset {
	// 创建 要素表
	if this.createFeaTable(name) {
		// 创建 要素集对象
		feaset := new(HbaseFeaset)
		feaset.store = this
		feaset.name = name
		feaset.geotype = geotype
		feaset.bbox = bbox
		feaset.count = 0
		feaset.index_level = HBASE_INDEX_LEVEL

		// 在系统表中增加一条记录；并且每次batch write之后，都应该更新系统表
		feaset.updateFeaTable()
		feaset.Open() // open还是默认打开吧 todo

		this.feasets = append(this.feasets, feaset)
		return feaset
	}
	feaset, _ := this.GetFeasetByName(name)
	return feaset
}

// 得到存储类型
func (this *HbaseStore) GetType() StoreType {
	return StoreHbase
}

// address
func (this *HbaseStore) GetConnParams() ConnParams {
	params := NewConnParams()
	params["address"] = this.address
	params["type"] = string(this.GetType())
	return params
}

func (this *HbaseStore) GetFeasetByNum(num int) (Featureset, error) {
	if num >= 0 && num < len(this.feasets) {
		return this.feasets[num], nil
	}
	return nil, errors.New("num must big than zero and less the count of feature sets.")
}

func (this *HbaseStore) GetFeasetByName(name string) (Featureset, error) {
	for _, v := range this.feasets {
		if v.name == name {
			return v, nil
		}
	}
	return nil, errors.New("cannot find the feature set of name: " + name + ".")
}

func (this *HbaseStore) GetFeasetNames() (names []string) {
	names = make([]string, len(this.feasets))
	for i, _ := range names {
		names[i] = this.feasets[i].name
	}
	return
}

// 关闭，释放资源
func (this *HbaseStore) Close() {
	this.hpool.Release()
}

type HbaseFeaset struct {
	name        string
	geotype     geometry.GeoType
	bbox        base.Rect2D
	count       int64
	index_level int32
	idx         index.ZOrderIndex
	store       *HbaseStore

	lock sync.Mutex // 并发写的时候，用来上锁
}

// 空间索引层级；先看看 5  层是否足够用
const HBASE_INDEX_LEVEL = 5

func (this *HbaseFeaset) Open() (bool, error) {
	// 似乎也就只能初始化空间索引了
	this.idx.InitDB(this.bbox, this.index_level)
	return true, nil
}

// 根据要素表的最新情况，更新系统表
// 在系统表中Put一条记录，包括：bbox、geotype、index_level,count 等信息，每次batch write之后，都应该更新信息
func (this *HbaseFeaset) updateFeaTable() {
	// client := gohbase.NewClient(this.store.address)
	client := this.store.openClient()
	value := map[string]map[string][]byte{
		"geo": { // 列族名, 与创建表时指定的名字保持一致
			// 列与值, 列名可自由定义
			"bbox":        this.bbox.ToBytes(),
			"geotype":     base.Int32ToBytes(int32(this.geotype)),
			"index_level": base.Int32ToBytes(HBASE_INDEX_LEVEL),
			"count":       base.Int64ToBytes(this.count),
		},
	}
	putRequest, _ := hrpc.NewPutStr(context.Background(), HBASE_SYS_TABLE, this.name, value)
	client.Put(putRequest)
	this.store.closeClient(client)
}

// 把索引编号和id合并为 []byte
func getRowkey(code int32, id int64) []byte {
	var buf = make([]byte, 12)
	// 这里必须用big，保证 code+1时，是在末端 增加 byte，而不是前端，保证row key的有序
	binary.BigEndian.PutUint32(buf, uint32(code))
	binary.BigEndian.PutUint64(buf[4:], uint64(id))
	return buf
}

func (this *HbaseFeaset) BatchWrite(feas []Feature) {
	// client := gohbase.NewClient(this.store.address)
	client := this.store.openClient()

	var bbox base.Rect2D
	bbox.Init()
	for _, v := range feas {
		code := this.idx.GetCode(v.Geo.GetBounds())
		rowkey := getRowkey(code, v.Geo.GetID())

		value := map[string]map[string][]byte{
			"geo": { // 列族名, 与创建表时指定的名字保持一致
				// 列与值, 列名可自由定义
				"geom": v.Geo.To(geometry.WKB), // 用wkb格式存储
			},
		}
		// fmt.Println("rowkey:", rowkey)
		putRequest, _ := hrpc.NewPut(context.Background(), []byte(this.name), rowkey, value)
		client.Put(putRequest)

		bbox.Union(v.Geo.GetBounds()) //
	}
	this.store.closeClient(client)

	this.lock.Lock()
	this.count += int64(len(feas))
	this.bbox.Union(bbox)
	this.lock.Unlock()
	return
}

// 结束批量写入，主要是更新系统表
func (this *HbaseFeaset) EndWrite() {
	this.updateFeaTable()
}

func (this *HbaseFeaset) Close() {
	return
}

func (this *HbaseFeaset) GetGeoType() geometry.GeoType {
	return this.geotype
}

func (this *HbaseFeaset) GetStore() Datastore {
	return this.store
}

func (this *HbaseFeaset) GetName() string {
	return this.name
}

// 对象个数
func (this *HbaseFeaset) GetCount() int64 {
	return this.count
}

func (this *HbaseFeaset) GetBounds() base.Rect2D {
	return this.bbox
}

// GetFieldInfos() []FieldInfo
// Query(bbox base.Rect2D, def QueryDef) FeatureIterator

func (this *HbaseFeaset) QueryByBounds(bbox base.Rect2D) FeatureIterator {
	itr := new(HbaseFeaItr)
	itr.feaset = this
	itr.bbox = bbox
	// 暂时没有找到快速统计 对象个数的办法，只好先按照面积平均分配了
	itr.count = int64(float64(this.count) * bbox.Area() / this.bbox.Area())
	itr.codes = this.idx.QueryDB(bbox) // codes 已排序
	return itr
}

// QueryByDef(def QueryDef) FeatureIterator

type HbaseFeaItr struct {
	feaset *HbaseFeaset
	bbox   base.Rect2D
	codes  []int32
	count  int64 // 对象数，猜测的

	countPerGo int       // 每一个批次的对象数量
	codess     [][]int32 // 每个批次所对应的index codes
}

func (this *HbaseFeaItr) Count() int64 {
	return this.count
}

// todo
// Next() (Feature, bool)

// 为调用批量读取做准备，调用 BatchNext 之前必须调用 本函数
// objCount 为每个批次拟获取对象的数量，不保证精确
func (this *HbaseFeaItr) PrepareBatch(objCount int) int {
	goCount := int(this.count)/objCount + 1
	// 这里假设每个code中所包含的对象，是大体平均分布的
	this.codess = base.SplitSlice32(this.codes, goCount)
	if len(this.codess) != goCount {
		fmt.Println("PrepareBatch error! len of codes:", this.codes, "go count:", goCount)
	}
	// fmt.Println("codes:", this.codes)
	// fmt.Println("codess:", this.codess)
	this.countPerGo = objCount
	return goCount
}

// 把数组是否连续进行分割
func buildNextSlices(in []int32) (outs [][]int32) {
	var out []int32
	for pos := 0; pos < len(in); pos++ {
		out = append(out, in[pos])
		if pos == len(in)-1 {
			// 已经到最后一个，加自己
			outs = append(outs, out)
		} else {
			// 自己不是最后一个，则看下一个是否连续
			if in[pos+1] != in[pos]+1 {
				// 不连续就中断
				outs = append(outs, out)
				out = make([]int32, 0) // 必须重新make，防止冲突
			}
		}
	}
	return
}

// func buildNextSlices2(in []int) (outs [][]int) {
// 	outs = make([][]int, len(in))
// 	for i, v := range in {
// 		outs[i] = make([]int, 1)
// 		outs[i][0] = v
// 	}
// 	return
// }

// 批量读取，支持go协程安全；调用前，务必调用 PrepareBatch
// batchNo 为批量的序号
// 只要读取到一个数据，达不到count的要求，也返回true
func (this *HbaseFeaItr) BatchNext2(batchNo int) (feas []Feature, result bool) {
	if batchNo < len(this.codess) {
		result = true // 只要no不越界，就返回 true
		codes := this.codess[batchNo]
		feas = make([]Feature, 0, this.countPerGo)

		codess := buildNextSlices(codes)
		// fmt.Println("BatchNext codess:", codess)

		var wg *sync.WaitGroup = new(sync.WaitGroup)
		feass := make([][]Feature, len(codess))
		for i, v := range codess {
			n := len(v) // 这个 v 中的id是连续的
			start := getRowkey(int32(v[0]), 0)
			end := getRowkey(int32(v[n-1]+1), 0)

			wg.Add(1)
			go this.batchNext(feass, i, start, end, wg, v[0], v[n-1])
		}
		wg.Wait()

		for _, v := range feass {
			feas = append(feas, v...)
		}
		// fmt.Println("feas count:", len(feas))
	}

	return
}

func (this *HbaseFeaItr) batchNext(feass [][]Feature, num int, start, end []byte, wg *sync.WaitGroup, code1, code2 int32) {
	defer wg.Done()

	// client := gohbase.NewClient(this.feaset.store.address)
	client := this.feaset.store.openClient()
	scanRequest, _ := hrpc.NewScanRange(context.Background(), []byte(this.feaset.name), start, end)
	scan := client.Scan(scanRequest)
	// count := 0
	for {
		var fea Feature
		getRsp, err := scan.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		if err != nil {
			fmt.Println("scan next error:", err)
		}
		for _, v := range getRsp.Cells {
			if string(v.Family) == "geo" && string(v.Qualifier) == "geom" {
				// fmt.Println("rowkey:", v.Row)
				fea.Geo = geometry.CreateGeo(this.feaset.geotype)
				fea.Geo.From(v.Value, geometry.WKB)
				// 用big，和构建row key保持一致
				id := int64(binary.BigEndian.Uint64(v.Row[4:]))
				fea.Geo.SetID(id)
				break
			}
		}
		// bbox相交，才取出去
		if this.bbox.IsIntersect(fea.Geo.GetBounds()) {
			// count++
			feass[num] = append(feass[num], fea)
		}
	}
	// fmt.Println("code1:", code1, "code2:", code2, "geo count:", len(feass[num]), "start:", start, "end:", end)
	// client.Close()
	this.feaset.store.closeClient(client)
}

func (this *HbaseFeaItr) BatchNext(batchNo int) (feas []Feature, result bool) {
	if batchNo < len(this.codess) {
		result = true // 只要no不越界，就返回 true
		codes := this.codess[batchNo]
		feas = make([]Feature, 0, this.countPerGo)

		codess := buildNextSlices(codes)

		for _, v := range codess {
			n := len(v) // 这个 v 中的id是连续的
			// client := gohbase.NewClient(this.feaset.store.address)
			client := this.feaset.store.openClient()
			start := getRowkey(int32(v[0]), 0)
			end := getRowkey(int32(v[n-1]+1), 0)
			scanRequest, _ := hrpc.NewScanRange(context.Background(), []byte(this.feaset.name), start, end)
			scan := client.Scan(scanRequest)
			count := 0
			// var rowkey []byte
			for {
				var fea Feature
				getRsp, err := scan.Next()
				if err == io.EOF || getRsp == nil {
					break
				}
				if err != nil {
					fmt.Println("scan next error:", err)
				}
				for _, vv := range getRsp.Cells {
					if string(vv.Family) == "geo" && string(vv.Qualifier) == "geom" {
						fea.Geo = geometry.CreateGeo(this.feaset.geotype)
						fea.Geo.From(vv.Value, geometry.WKB)
						// 用big，和构建row key保持一致
						id := int64(binary.BigEndian.Uint64(vv.Row[4:]))
						fea.Geo.SetID(id)
						count++
						break
					}
				}
				// bbox相交，才取出去
				if this.bbox.IsIntersect(fea.Geo.GetBounds()) {
					feas = append(feas, fea)
				}
			}
			// client.Close()
			this.feaset.store.closeClient(client)
		}
	}

	return
}

func (this *HbaseFeaItr) Close() {
	this.codes = this.codes[:0]
	this.codess = this.codess[:0]
}
