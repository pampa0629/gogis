package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/geometry"
	"math"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
)

type EsFeaset struct {
	store *EsStore
	data.FeasetInfo
	// name    string
	// bbox    base.Rect2D
	// geotype geometry.GeoType
	count int64
	// data.ProjCommon
}

func (this *EsFeaset) getCount() {
	config := elasticsearch.Config{
		Addresses: this.store.addresses,
	}
	es, _ := elasticsearch.NewClient(config)
	res, _ := es.Cat.Count(es.Cat.Count.WithIndex(this.Name))
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
		es.Search.WithIndex(this.Name),
		es.Search.WithBody(&buf),
		es.Search.WithSize(0))
	// fmt.Println("get bbox, res:", res)
	maps := esRes2maps(res)
	maxy := getMapsValue(maps, "aggregations", "viewport", "bounds", "top_left", "lat").(float64)
	minx := getMapsValue(maps, "aggregations", "viewport", "bounds", "top_left", "lon").(float64)
	miny := getMapsValue(maps, "aggregations", "viewport", "bounds", "bottom_right", "lat").(float64)
	maxx := getMapsValue(maps, "aggregations", "viewport", "bounds", "bottom_right", "lon").(float64)
	this.Bbox = base.NewRect2D(minx, miny, maxx, maxy)
}

// 内部调用，和Open做区分
func (this *EsFeaset) open() {
	this.Proj = base.PrjFromEpsg(4326)
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

func (this *EsFeaset) GetStore() data.Datastore {
	return this.store
}

// func (this *EsFeaset) GetName() string {
// 	return this.name
// }

func (this *EsFeaset) GetCount() int64 {
	return this.count
}

// func (this *EsFeaset) GetBounds() base.Rect2D {
// 	return this.bbox
// }

// func (this *EsFeaset) GetGeoType() geometry.GeoType {
// 	return this.geotype
// }

// todo
func (this *EsFeaset) BeforeWrite(count int64) {
}

// 批量写入
func (this *EsFeaset) BatchWrite(feas []data.Feature) {
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

	res, err := es.Bulk(bytes.NewReader(buf.Bytes()), es.Bulk.WithIndex(this.Name))
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

// todo
func (this *EsFeaset) Query(def *data.QueryDef) data.FeatureIterator {
	if def == nil {
		def = data.NewQueryDef(this.Bbox)
	}
	bbox := def.SpatialObj.(base.Rect2D)
	// todo 其他查询条件的响应
	return this.queryByBounds(bbox)
}

func (this *EsFeaset) queryByBounds(bbox base.Rect2D) data.FeatureIterator {
	bbox = limitGlobal(bbox)
	res := esQueryWithBounds(this.store.addresses, this.Name, bbox, 0, 0)
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
