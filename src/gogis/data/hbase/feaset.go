package hbase

import (
	"context"
	"encoding/binary"
	"gogis/base"
	"gogis/data"
	"gogis/geometry"
	"gogis/index"
	"sync"

	"github.com/tsuna/gohbase/hrpc"
)

type HbaseFeaset struct {
	// name        string
	// geotype     geometry.GeoType
	// bbox        base.Rect2D
	data.FeasetInfo
	count       int64
	index_level int32
	idx         index.ZOrderIndex
	store       *HbaseStore
	// data.ProjCommon

	lock sync.Mutex // 并发写的时候，用来上锁
}

// 空间索引层级；先看看 5  层是否足够用
const HBASE_INDEX_LEVEL = 5

func (this *HbaseFeaset) Open() (bool, error) {
	// 似乎也就只能初始化空间索引了
	this.idx.InitDB(this.Bbox, this.index_level)
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
			"bbox":        this.Bbox.ToBytes(),
			"geotype":     base.Int32ToBytes(int32(this.GeoType)),
			"index_level": base.Int32ToBytes(HBASE_INDEX_LEVEL),
			"count":       base.Int64ToBytes(this.count),
		},
	}
	putRequest, _ := hrpc.NewPutStr(context.Background(), HBASE_SYS_TABLE, this.Name, value)
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

// todo
func (this *HbaseFeaset) BeforeWrite(count int64) {
}

func (this *HbaseFeaset) BatchWrite(feas []data.Feature) {
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
		putRequest, _ := hrpc.NewPut(context.Background(), []byte(this.Name), rowkey, value)
		client.Put(putRequest)

		bbox = bbox.Union(v.Geo.GetBounds()) //
	}
	this.store.closeClient(client)

	this.lock.Lock()
	this.count += int64(len(feas))
	this.Bbox = this.Bbox.Union(bbox)
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

func (this *HbaseFeaset) GetStore() data.Datastore {
	return this.store
}

// 对象个数
func (this *HbaseFeaset) GetCount() int64 {
	return this.count
}

func (this *HbaseFeaset) queryByBounds(bbox base.Rect2D) data.FeatureIterator {
	itr := new(HbaseFeaItr)
	itr.feaset = this
	itr.bbox = bbox
	// 暂时没有找到快速统计 对象个数的办法，只好先按照面积平均分配了
	itr.count = int64(float64(this.count) * bbox.Area() / this.Bbox.Area())
	itr.codes = this.idx.QueryDB(bbox) // codes 已排序
	return itr
}

func (this *HbaseFeaset) Query(def *data.QueryDef) data.FeatureIterator {
	if def == nil {
		def = data.NewQueryDef(this.Bbox)
	}
	bbox := def.SpatialObj.(base.Rect2D)
	return this.queryByBounds(bbox)
}
