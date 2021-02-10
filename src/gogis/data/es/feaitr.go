package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/geometry"
	"strconv"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
)

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
func (this *EsFeaItr) BeforeNext(objCount int) int {
	// es 默认每次最多取10000个对象，这里控制一下
	this.countPerBatch = base.IntMin(objCount, 10000)
	goCount := int(this.count/int64(this.countPerBatch) + 1)
	this.froms = make([]int, goCount)
	for i, _ := range this.froms {
		this.froms[i] = i * this.countPerBatch
	}
	return goCount
}

// 批量读取，支持go协程安全；调用前，务必调用 BeforeNext
// batchNo 为批量的序号
// 只要读取到一个数据，达不到count的要求，也返回true
func (this *EsFeaItr) BatchNext(batchNo int) ([]data.Feature, bool) {
	if batchNo > len(this.froms) {
		return nil, false
	}

	count := this.countPerBatch // 这个批次取多少个对象
	if int64((batchNo+1)*count) > this.count {
		// 如果剩余数量不够，就取剩余的
		count = int(this.count - int64(batchNo*count))
	}

	res := esQueryWithBounds(this.feaset.store.addresses, this.feaset.Name, this.bbox, this.froms[batchNo], count)
	// fmt.Println("batch next, res:", res)
	maps := esRes2maps(res)
	values := getMapsValue(maps, "hits", "hits")
	pnts := values.([]interface{})
	feas := make([]data.Feature, 0)
	for _, v := range pnts {
		var fea data.Feature
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
	res := esAggGrids(this.feaset.store.addresses, this.feaset.Name, this.bbox, precision)
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
