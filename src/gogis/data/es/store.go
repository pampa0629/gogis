package es

import (
	"bytes"
	"context"
	"encoding/json"
	"gogis/data"
	"gogis/geometry"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
)

func init() {
	data.RegisterDatastore(data.StoreES, NewEsStore)
}

func NewEsStore() data.Datastore {
	return new(EsStore)
}

// elasticsearc 引擎，优先存储点数据，做网格聚合图等
// 只存储经纬度数据
type EsStore struct {
	addresses []string // es address, such as: localhost:9200 or ip:9201
	data.Featuresets
	// hpool pool.Pool
}

func (this *EsStore) Open(params data.ConnParams) (bool, error) {
	this.addresses = this.addresses[:0]      // 清空原来的
	if temp, ok := params["addresses"]; ok { // 有
		if address, ok := temp.(string); ok { // 是字符串类型
			// 兼容只输入一个字符串
			this.addresses = []string{address}
		} else if addresses, ok := temp.([]interface{}); ok { // 是数组
			for _, v := range addresses {
				this.addresses = append(this.addresses, v.(string))
			}
		}
	}

	config := elasticsearch.Config{
		Addresses: this.addresses,
	}
	// new client
	es, err := elasticsearch.NewClient(config)
	if err == nil {
		// 这里应该要得到 数据集的系统信息
		indices := this.getIndexNames(es)
		for _, v := range indices {
			indextype := this.getIndexType(es, v)
			if indextype == "geo_point" {
				feaset := new(EsFeaset)
				// open或者get时，再提供
				// count   int64
				// bbox    base.Rect2D
				feaset.store = this
				feaset.Name = v
				feaset.GeoType = geometry.TGeoPoint
				this.Feasets = append(this.Feasets, feaset)
			}
		}
		return true, nil
	}
	return false, err
}

func getMapsValue(maps map[string]interface{}, keys ...string) (value interface{}) {
	lens := len(keys)
	for i := 0; i < lens-1; i++ {
		temp, ok := maps[keys[i]].(map[string]interface{})
		if ok && temp != nil {
			maps = temp
		} else {
			break
		}
	}
	return maps[keys[lens-1]]
}

// 得到index的类型，看是否为 空间类型
func (this *EsStore) getIndexType(es *elasticsearch.Client, index string) (indextype string) {
	res, err := es.Indices.GetMapping(es.Indices.GetMapping.WithIndex(index))
	if err == nil {
		maps := esRes2maps(res)
		value := getMapsValue(maps, index, "mappings", "properties", "location", "type")
		if value != nil {
			indextype = value.(string)
		}
	}
	defer res.Body.Close()
	return
}

func esRes2maps(res *esapi.Response) map[string]interface{} {
	var maps map[string]interface{}
	bytes := make([]byte, len(res.String()))
	n, _ := res.Body.Read(bytes)
	json.Unmarshal(bytes[0:n], &maps)
	return maps
}

// 得到所有 index的名单，包括非空间数据
func (this *EsStore) getIndexNames(es *elasticsearch.Client) (names []string) {
	res, err := es.Indices.GetAlias(es.Indices.GetAlias.WithContext(context.Background()))
	if err == nil {
		maps := esRes2maps(res)
		for k, _ := range maps {
			names = append(names, k)
		}
	}
	defer res.Body.Close()
	return
}

func (this *EsStore) GetType() data.StoreType {
	return data.StoreES
}

func (this *EsStore) GetConnParams() data.ConnParams {
	params := data.NewConnParams()
	params["addresses"] = this.addresses
	params["type"] = string(this.GetType())
	params["gowrite"] = 100 // todo
	return params
}

func (this *EsStore) Close() {

}

// 创建es中的index
func (this *EsStore) createIndex(name string) bool {
	config := elasticsearch.Config{
		Addresses: this.addresses,
	}
	es, _ := elasticsearch.NewClient(config)
	var buf bytes.Buffer
	mappings := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"gid": map[string]interface{}{
					"type": "long",
				},
				"location": map[string]interface{}{
					"type": "geo_point",
				},
			},
		},
	}
	json.NewEncoder(&buf).Encode(mappings)

	res, err := es.Indices.Create(name, es.Indices.Create.WithBody(&buf))
	defer res.Body.Close()
	return err == nil
}

// 创建要素集
func (this *EsStore) CreateFeaset(info data.FeasetInfo) data.Featureset {
	// todo 暂且只支持点数据
	if info.GeoType == geometry.TGeoPoint {
		if this.createIndex(info.Name) {
			feaset := new(EsFeaset)
			feaset.store = this
			feaset.FeasetInfo = info
			// feaset.name = name
			// feaset.geotype = geotype
			feaset.Bbox.Init()
			feaset.open()
			this.Feasets = append(this.Feasets, feaset)
			return feaset
		}
	}
	return nil
}

// todo
func (this *EsStore) DeleteFeaset(name string) bool {
	return false
}
