package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gogis/base"
	"log"
	"sync"

	//"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch"
)

func esmain() {
	tr := base.NewTimeRecorder()
	// Create()
	// Doc()
	// Get()

	Search(0, nil)
	// count := 2600
	// var wg *sync.WaitGroup = new(sync.WaitGroup)
	// for i := 0; i < count; i++ {
	// 	wg.Add(1)
	// 	go Search(i, wg)
	// }
	// wg.Wait()

	// Delete()
	tr.Output("main")
	fmt.Println("DONE!")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func Create() {
	addresses := []string{"http://127.0.0.1:9200"}
	config := elasticsearch.Config{
		Addresses: addresses,
	}
	// new client
	es, err := elasticsearch.NewClient(config)
	failOnError(err, "Error creating the client")
	// Index creates or updates a document in an index
	var buf bytes.Buffer
	mappings := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"location": map[string]interface{}{
					"type": "geo_point",
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(mappings); err != nil {
		failOnError(err, "Error encoding doc")
	}
	res, err := es.Indices.Create("geodata", es.Indices.Create.WithBody(&buf))

	if err != nil {
		failOnError(err, "Error Index response")
	}
	defer res.Body.Close()
	fmt.Println(res.String())
}

func Doc() {
	addresses := []string{"http://127.0.0.1:9200"}
	config := elasticsearch.Config{
		Addresses: addresses,
		Username:  "",
		Password:  "",
		CloudID:   "",
		APIKey:    "",
	}
	// new client
	es, err := elasticsearch.NewClient(config)
	failOnError(err, "Error creating the client")
	// Index creates or updates a document in an index
	var buf bytes.Buffer
	data := map[string]interface{}{
		"doc_id": "123",
		"location": map[string]interface{}{
			"lat": 40.822,
			"lon": 73.889,
		},
	}

	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		failOnError(err, "Error encoding doc")
	}

	// res, err := es.Index("geodata", &buf, es.Index.WithDocumentID("1"))
	res, err := es.Index("geodata", &buf)
	if err != nil {
		failOnError(err, "Error Index response")
	}
	defer res.Body.Close()
	fmt.Println(res.String())
}

func Search(n int, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	addresses := []string{"http://127.0.0.1:9200", "http://127.0.0.1:9201"}
	config := elasticsearch.Config{
		Addresses: addresses,
	}
	// new client
	es, err := elasticsearch.NewClient(config)
	failOnError(err, "Error creating the client")

	// info
	// res, err := es.Info()
	name := "insurance"

	// search - highlight
	var buf bytes.Buffer
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"constant_score": map[string]interface{}{
				"filter": map[string]interface{}{
					"geo_bounding_box": map[string]interface{}{
						"location": map[string]interface{}{
							"top_left": map[string]interface{}{
								"lat": 90,
								"lon": 10,
							},
							"bottom_right": map[string]interface{}{
								"lat": 0,
								"lon": 170,
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
					"precision": 4,
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
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		failOnError(err, "Error encoding query")
	}
	// Perform the search request.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(name),
		es.Search.WithBody(&buf),
		// es.Search.WithScroll(time.Second*10),
		// es.Search.WithSize(100), // 默认配置，最大1万
		// es.Search.WithFrom(12000),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)

	if err != nil {
		failOnError(err, "Error getting response")
	}
	defer res.Body.Close()
	str := res.String()
	// es.Search.()
	fmt.Println("len:", len(str))
	fmt.Println("res:", str)
	// fmt.Println("n:", n)
}

type Mappings map[string]interface{}

func Get() {
	addresses := []string{"http://127.0.0.1:9200", "http://127.0.0.1:9201"}
	// config := elasticsearch.Config{
	// 	Addresses: addresses,
	// 	Username:  "",
	// 	Password:  "",
	// 	CloudID:   "",
	// 	APIKey:    "",
	// }
	config := elasticsearch.Config{
		Addresses: addresses,
	}
	// new client
	es, err := elasticsearch.NewClient(config)
	failOnError(err, "Error creating the client")
	// res, err := es.Indices.GetAlias(es.Indices.GetAlias.WithContext(context.Background()))
	// res, err := es.Indices.Stats(es.Indices.Stats.WithContext(context.Background()))
	// res, err := es.Indices.GetMapping(es.Indices.GetMapping.WithIndex("geodata"))
	// failOnError(err, "Error Indices.Stats")
	// var maps map[string]interface{}
	// bytes := make([]byte, len(res.String()))
	// n, e := res.Body.Read(bytes)
	// fmt.Println("bytes:", n, e, string(bytes))
	// json.Unmarshal(bytes[0:n], &maps)
	// fmt.Println("maps:", maps)

	res, err := es.Get("geodata", "Zch40XYBAEPGr4siwOM-")
	if err != nil {
		failOnError(err, "Error get response")
	}
	defer res.Body.Close()
	fmt.Println(res.String())
}

func Delete() {
	addresses := []string{"http://127.0.0.1:9200", "http://127.0.0.1:9201"}
	config := elasticsearch.Config{
		Addresses: addresses,
		Username:  "",
		Password:  "",
		CloudID:   "",
		APIKey:    "",
	}
	// new client
	es, err := elasticsearch.NewClient(config)
	failOnError(err, "Error creating the client")
	res, err := es.Indices.Delete([]string{"geodata"})
	// res, err := es.Get("geodata", "1")
	if err != nil {
		failOnError(err, "Error get response")
	}
	defer res.Body.Close()
	fmt.Println(res.String())
}
