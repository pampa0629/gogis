package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"math"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
)

// elasticsearc 引擎，优先存储点数据，做网格聚合图等
// 只存储经纬度数据
type EsStore struct {
	addresses []string // es address, such as: localhost:9200 or ip:9201
	Feasets
	// hpool pool.Pool
}

func (this *EsStore) Open(params ConnParams) (bool, error) {
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
				feaset.name = v
				feaset.geotype = geometry.TGeoPoint
				this.feasets = append(this.feasets, feaset)
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

func (this *EsStore) GetType() StoreType {
	return StoreES
}

func (this *EsStore) GetConnParams() ConnParams {
	params := NewConnParams()
	params["addresses"] = this.addresses
	params["type"] = string(this.GetType())
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
func (this *EsStore) CreateFeaset(name string, bbox base.Rect2D, geotype geometry.GeoType) Featureset {
	// todo 暂且只支持点数据
	if geotype == geometry.TGeoPoint {
		if this.createIndex(name) {
			feaset := new(EsFeaset)
			feaset.store = this
			feaset.name = name
			feaset.geotype = geotype
			feaset.bbox.Init()
			feaset.open()
			this.feasets = append(this.feasets, feaset)
			return feaset
		}
	}
	return nil
}

type EsFeaset struct {
	store   *EsStore
	name    string
	count   int64
	bbox    base.Rect2D
	geotype geometry.GeoType
	projCommon
}

func (this *EsFeaset) getCount() {
	config := elasticsearch.Config{
		Addresses: this.store.addresses,
	}
	es, _ := elasticsearch.NewClient(config)
	res, _ := es.Cat.Count(es.Cat.Count.WithIndex(this.name))
	// res 最后的信息为 count
	items := strings.Split(res.String(), " ")
	strCount := strings.Replace(items[len(items)-1], "\n", "", -1)
	this.count, _ = strconv.ParseInt(strCount, 10, 64)
	defer res.Body.Close()
}

func (this *EsFeaset) getBbox() {
	config := elasticsearch.Config{
		Addresses: this.store.addresses,
	}
	es, _ := elasticsearch.NewClient(config)

	var buf bytes.Buffer
	query := map[string]interface{}{
		"aggs": map[string]interface{}{
			"viewport": map[string]interface{}{
				"geo_bounds": map[string]interface{}{
					"field": "location"}},
		},
	}
	json.NewEncoder(&buf).Encode(query)

	res, _ := es.Search(
		es.Search.WithIndex(this.name),
		es.Search.WithBody(&buf),
		es.Search.WithSize(0))
	// fmt.Println("get bbox, res:", res)
	maps := esRes2maps(res)
	maxy := getMapsValue(maps, "aggregations", "viewport", "bounds", "top_left", "lat").(float64)
	minx := getMapsValue(maps, "aggregations", "viewport", "bounds", "top_left", "lon").(float64)
	miny := getMapsValue(maps, "aggregations", "viewport", "bounds", "bottom_right", "lat").(float64)
	maxx := getMapsValue(maps, "aggregations", "viewport", "bounds", "bottom_right", "lon").(float64)
	this.bbox = base.NewRect2D(minx, miny, maxx, maxy)
}

// 内部调用，和Open做区分
func (this *EsFeaset) open() {
	this.proj = base.PrjFromEpsg(4326)
}

func (this *EsFeaset) Open() (bool, error) {
	// 这里统计 count和bbox（通过 geo_bounds ）
	this.getCount()
	this.getBbox()
	this.open()
	return true, nil
}

func (this *EsFeaset) Close() {

}

func (this *EsFeaset) GetStore() Datastore {
	return this.store
}

func (this *EsFeaset) GetName() string {
	return this.name
}

func (this *EsFeaset) GetCount() int64 {
	return this.count
}

func (this *EsFeaset) GetBounds() base.Rect2D {
	return this.bbox
}

func (this *EsFeaset) GetGeoType() geometry.GeoType {
	return this.geotype
}

// 批量写入
func (this *EsFeaset) BatchWrite(feas []Feature) {
	config := elasticsearch.Config{
		Addresses: this.store.addresses,
	}
	es, _ := elasticsearch.NewClient(config)

	var buf bytes.Buffer
	for _, v := range feas {
		pnt := v.Geo.(*geometry.GeoPoint)
		geo := map[string]interface{}{
			"gid": v.Geo.GetID(),
			"location": map[string]interface{}{
				"lat": pnt.Y,
				"lon": pnt.X,
			},
		}
		// 必须有 meta，才能正确写入
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%d" } }%s`, v.Geo.GetID(), "\n"))
		data, _ := json.Marshal(geo)
		data = append(data, "\n"...)
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}

	res, err := es.Bulk(bytes.NewReader(buf.Bytes()), es.Bulk.WithIndex(this.name))
	if err != nil {
		fmt.Println("es bulk err:", err)
	}
	defer res.Body.Close()
}

// todo 貌似无事可做
func (this *EsFeaset) EndWrite() {
}

// 把bbox限定在全球经纬度范围内
func limitGlobal(bbox base.Rect2D) base.Rect2D {
	bbox.Min.X = math.Max(bbox.Min.X, -180)
	bbox.Min.Y = math.Max(bbox.Min.Y, -90)
	bbox.Max.X = math.Min(bbox.Max.X, 180)
	bbox.Max.Y = math.Min(bbox.Max.Y, 90)
	return bbox
}

func (this *EsFeaset) QueryByBounds(bbox base.Rect2D) FeatureIterator {
	bbox = limitGlobal(bbox)
	res := esQueryWithBounds(this.store.addresses, this.name, bbox, 0, 0)
	if res.StatusCode < 400 {
		maps := esRes2maps(res)
		defer res.Body.Close()

		feaItr := new(EsFeaItr)
		feaItr.bbox = bbox
		feaItr.count = int64(getMapsValue(maps, "hits", "total", "value").(float64))
		feaItr.feaset = this
		return feaItr
	}
	fmt.Println("es feaset query by bounds error:", res.String())
	return nil
}

func esQueryWithBounds(addresses []string, name string, bbox base.Rect2D, from, size int) *esapi.Response {
	config := elasticsearch.Config{
		Addresses: addresses,
	}
	es, _ := elasticsearch.NewClient(config)

	var buf bytes.Buffer
	query := map[string]interface{}{
		"search_after": []int{from - 1},
		"sort": map[string]interface{}{
			"gid": "asc"},
		"query": map[string]interface{}{
			"geo_bounding_box": map[string]interface{}{
				"location": map[string]interface{}{
					"top_left": map[string]interface{}{
						"lat": bbox.Max.Y,
						"lon": bbox.Min.X},
					"bottom_right": map[string]interface{}{
						"lat": bbox.Min.Y,
						"lon": bbox.Max.X}}}}}
	json.NewEncoder(&buf).Encode(query)

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(name),
		es.Search.WithBody(&buf),
		es.Search.WithSize(size), // 默认配置，最大1万
		// es.Search.WithFrom(from),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err == nil {
		return res
	}
	return nil
}

type EsFeaItr struct {
	feaset        *EsFeaset
	bbox          base.Rect2D
	count         int64
	froms         []int
	countPerBatch int // 每个批次的对象个数
}

func (this *EsFeaItr) Count() int64 {
	return this.count
}

// 为调用批量读取做准备，调用 BatchNext 之前必须调用 本函数
// objCount 为每个批次拟获取对象的数量，不保证精确
func (this *EsFeaItr) PrepareBatch(objCount int) int {
	// es 默认每次最多取10000个对象，这里控制一下
	this.countPerBatch = base.IntMin(objCount, 10000)
	goCount := int(this.count/int64(this.countPerBatch) + 1)
	this.froms = make([]int, goCount)
	for i, _ := range this.froms {
		this.froms[i] = i * this.countPerBatch
	}
	return goCount
}

// 批量读取，支持go协程安全；调用前，务必调用 PrepareBatch
// batchNo 为批量的序号
// 只要读取到一个数据，达不到count的要求，也返回true
func (this *EsFeaItr) BatchNext(batchNo int) ([]Feature, bool) {
	if batchNo > len(this.froms) {
		return nil, false
	}

	count := this.countPerBatch // 这个批次取多少个对象
	if int64((batchNo+1)*count) > this.count {
		// 如果剩余数量不够，就取剩余的
		count = int(this.count - int64(batchNo*count))
	}

	res := esQueryWithBounds(this.feaset.store.addresses, this.feaset.name, this.bbox, this.froms[batchNo], count)
	// fmt.Println("batch next, res:", res)
	maps := esRes2maps(res)
	values := getMapsValue(maps, "hits", "hits")
	pnts := values.([]interface{})
	feas := make([]Feature, 0)
	for _, v := range pnts {
		var fea Feature
		geoPoint := new(geometry.GeoPoint)
		pntMap := v.(map[string]interface{})
		id, _ := strconv.Atoi(getMapsValue(pntMap, "_id").(string))
		geoPoint.SetID(int64((id)))
		geoPoint.Y = getMapsValue(pntMap, "_source", "location", "lat").(float64)
		geoPoint.X = getMapsValue(pntMap, "_source", "location", "lon").(float64)
		// fea.Geo = geometry.CreateGeo(this.feaset.geotype)
		fea.Geo = geoPoint
		feas = append(feas, fea)
	}
	defer res.Body.Close()
	return feas, true
}

func esAggGrids(addresses []string, name string, bbox base.Rect2D, precision int) *esapi.Response {
	config := elasticsearch.Config{
		Addresses: addresses,
	}
	es, err := elasticsearch.NewClient(config)

	var buf bytes.Buffer
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"constant_score": map[string]interface{}{
				"filter": map[string]interface{}{
					"geo_bounding_box": map[string]interface{}{
						"location": map[string]interface{}{
							"top_left": map[string]interface{}{
								"lat": bbox.Max.Y,
								"lon": bbox.Min.X},
							"bottom_right": map[string]interface{}{
								"lat": bbox.Min.Y,
								"lon": bbox.Max.X,
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			name: map[string]interface{}{
				"geohash_grid": map[string]interface{}{
					"field":     "location",
					"precision": precision,
				},
				"aggs": map[string]interface{}{
					"cell": map[string]interface{}{
						"geo_bounds": map[string]interface{}{
							"field": "location",
						},
					},
				},
			},
		},
	}
	json.NewEncoder(&buf).Encode(query)

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(name),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("agg grid, error:", err)
	}
	// str := res.String()
	return res
}

// 得到网格聚合图
func (this *EsFeaItr) AggGrids(precision int) (bboxes []base.Rect2D, counts []int) {
	res := esAggGrids(this.feaset.store.addresses, this.feaset.name, this.bbox, precision)
	maps := esRes2maps(res)
	buckets := getMapsValue(maps, "aggregations", "insurance", "buckets")
	values := buckets.([]interface{})
	for _, vv := range values {
		v := vv.(map[string]interface{})
		count := getMapsValue(v, "doc_count")
		counts = append(counts, int(count.(float64)))
		var bbox base.Rect2D
		bbox.Max.Y = getMapsValue(v, "cell", "bounds", "top_left", "lat").(float64)
		bbox.Min.X = getMapsValue(v, "cell", "bounds", "top_left", "lon").(float64)
		bbox.Min.Y = getMapsValue(v, "cell", "bounds", "bottom_right", "lat").(float64)
		bbox.Max.X = getMapsValue(v, "cell", "bounds", "bottom_right", "lon").(float64)
		bboxes = append(bboxes, bbox)
	}
	defer res.Body.Close()
	return
}

// 关闭，释放资源
func (this *EsFeaItr) Close() {

}
