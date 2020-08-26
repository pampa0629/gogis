package data

import (
	"gogis/base"
	"gogis/geometry"
)

// 打开数据的连接参数
type ConnParams map[string]string

// type Connection struct {
// 	server   string
// 	user     string
// 	password string
// }

// 数据存储库
type Datastore interface {
	Open(params ConnParams) (bool, error)
	GetFeatureset(name string) (Featureset, error)
	FeaturesetNames() []string
	Close() // 关闭，释放资源
}

// 矢量数据集合
type Featureset interface {
	Open(name string) (bool, error)
	GetName() string
	GetBounds() base.Rect2D
	Query(bbox base.Rect2D) FeatureIterator
	Close()
}

// 集合对象迭代器，用来遍历对象
type FeatureIterator interface {
	// HasNext() bool
	Next() (Feature, bool)
}

// 一个矢量对象（带属性）
type Feature struct {
	geo    geometry.Geometry
	fields map[string]interface{}
}

// 栅格数据集合 todo
type Rasterset interface {
	GetBounds() base.Rect2D
}
