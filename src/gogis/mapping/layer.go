package mapping

import (
	"encoding/json"
	"gogis/base"
	"gogis/data"
	"gogis/draw"
	"strings"
)

// 图层类型定义
type LayerType string

const (
	LayerFeature LayerType = "FeatureLayer" // 要素图层
	LayerRaster  LayerType = "RasterLayer"  // 栅格图层
)

// 图层类接口
type Layer interface {
	GetBounds() base.Rect2D        // base.Bounds
	GetProjection() *base.ProjInfo // 得到投影坐标系，没有返回nil
	GetName() string
	GetConnParams() data.ConnParams
	GetType() LayerType

	Draw(canvas draw.Canvas, proj *base.ProjInfo) int64

	Close()
	WhenSaving(mappath string)
	WhenOpening(mappath string)
}

// type Layers struct {
// 	[]Layer
// }

type Layers []Layer

func createLayer(layerType LayerType, jsonValue string) (layer Layer) {
	switch layerType {
	case LayerFeature:
		layer = new(FeatureLayer)
	case LayerRaster:
		layer = new(RasterLayer)
	}
	json.Unmarshal([]byte(jsonValue), layer)
	return
}

func (this *Layers) UnmarshalJSON(data []byte) error {
	jsons := splitJsons(string(data))
	for _, v := range jsons {
		values := strings.Split(v, "\n")
		var layerType LayerType
		for _, value := range values {
			if strings.Index(value, `"LayerType"`) >= 0 {
				if strings.Index(value, string(LayerRaster)) >= 0 {
					layerType = LayerRaster
				} else if strings.Index(value, string(LayerFeature)) >= 0 {
					layerType = LayerFeature
				}
				break
			}
		}
		layer := createLayer(layerType, v)
		*this = append(*this, layer)
	}
	return nil
}

// 根据 {} 划分json数组
func splitJsons(json string) (jsons []string) {
	indent := 0
	start := 0
	for i, v := range json {
		switch v {
		case '{':
			if indent == 0 {
				start = i
			}
			indent++
		case '}':
			indent--
			if indent == 0 {
				one := json[start : i+1]
				jsons = append(jsons, one)
				start = i + 1
			}
		}
	}
	return
}
