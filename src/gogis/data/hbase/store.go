package hbase

import (
	"context"
	"errors"
	"gogis/base"
	"gogis/data"
	"gogis/geometry"
	"io"
	"time"

	pool "github.com/silenceper/poor"
	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"
)

func init() {
	data.RegisterDatastore(data.StoreHbase, NewHbaseStore)
}

func NewHbaseStore() data.Datastore {
	return new(HbaseStore)
}

type HbaseStore struct {
	address string // zookeeper's address, such as: localhost:2182 or ip:2181
	// feasets []*HbaseFeaset
	data.Featuresets
	hpool pool.Pool
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

func (this *HbaseStore) Open(params data.ConnParams) (bool, error) {
	this.address = params["address"].(string)
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

	this.Feasets = make([]data.Featureset, 0)
	for {
		getRsp, err := scan.Next()
		if err == io.EOF || getRsp == nil {
			break
		}
		feaset := new(HbaseFeaset)
		feaset.store = this
		for _, v := range getRsp.Cells {
			feaset.Name = string(v.Row)
			// "geo": { // 列族名, 与创建表时指定的名字保持一致
			// 	"bbox":        this.bbox.ToBytes(),
			// 	"geotype":     base.Int32ToBytes(int32(this.geotype)),
			// 	"index_level": base.Int32ToBytes(HBASE_INDEX_LEVEL),
			// 	"count":       base.Int32ToBytes(int32(this.count)),
			// },
			if string(v.Family) == "geo" {
				if string(v.Qualifier) == "bbox" {
					feaset.Bbox.FromBytes(v.Value)
				} else if string(v.Qualifier) == "geotype" {
					feaset.GeoType = geometry.GeoType(base.BytesToInt32(v.Value))
				} else if string(v.Qualifier) == "index_level" {
					feaset.index_level = base.BytesToInt32(v.Value)
				} else if string(v.Qualifier) == "count" {
					feaset.count = base.BytesToInt64(v.Value)
				}
			}
		}
		this.Feasets = append(this.Feasets, feaset)
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
func (this *HbaseStore) DeleteFeaset(name string) bool {
	feaset, _ := this.GetFeasetByName(name)
	if feaset != nil {
		// 删除 要素表
		this.deleteFeaTable(name)

		// 删除系统表中的相关记录
		this.deleteSysInfo(name)

		// 从[]中删除
		for i, v := range this.Feasets {
			if v == feaset {
				this.Feasets = append(this.Feasets[:i], this.Feasets[i+1:]...)
				break
			}
		}
		return true
	}
	return false // , errors.New("feaset name: " + name + " is not existed.")
}

// 创建要素集
func (this *HbaseStore) CreateFeaset(info data.FeasetInfo) data.Featureset {
	// 创建 要素表
	if this.createFeaTable(info.Name) {
		// 创建 要素集对象
		feaset := new(HbaseFeaset)
		feaset.store = this
		feaset.FeasetInfo = info
		// feaset.name = name
		// feaset.geotype = geotype
		// feaset.bbox = bbox
		feaset.count = 0
		feaset.index_level = HBASE_INDEX_LEVEL

		// 在系统表中增加一条记录；并且每次batch write之后，都应该更新系统表
		feaset.updateFeaTable()
		feaset.Open() // open还是默认打开吧 todo

		this.Feasets = append(this.Feasets, feaset)
		return feaset
	}
	feaset, _ := this.GetFeasetByName(info.Name)
	return feaset
}

// 得到存储类型
func (this *HbaseStore) GetType() data.StoreType {
	return data.StoreHbase
}

// address
func (this *HbaseStore) GetConnParams() data.ConnParams {
	params := data.NewConnParams()
	params["address"] = this.address
	params["type"] = string(this.GetType())
	params["gowrite"] = 100 // todo
	return params
}

// 关闭，释放资源
func (this *HbaseStore) Close() {
	this.hpool.Release()
}
