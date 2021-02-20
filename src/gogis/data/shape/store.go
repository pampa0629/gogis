// shape存储库，硬盘模式
package shape

import (
	"gogis/base"
	"gogis/data"
)

func init() {
	data.RegisterDatastore(data.StoreShape, NewShapeStore)
}

func NewShapeStore() data.Datastore {
	return new(ShapeStore)
}

// // 快捷方法，打开一个shape文件，得到要素集对象
func OpenShape(filename string, cache bool, fields []string) data.Featureset {
	// 默认用内存模式
	// shp := new(ShpmemStore)
	shp := new(ShapeStore)
	params := data.NewConnParams()
	params["filename"] = filename
	params["cache"] = cache
	params["fields"] = fields
	shp.Open(params)
	feaset := shp.GetFeasetByNum(0)
	feaset.Open()
	// if cache {
	// 	return data.Cache(feaset, fields)
	// }
	return feaset
}

// todo 未来还要考虑实现打开一个文件夹
// 硬盘模式的shape存储库
type ShapeStore struct {
	feaset   *ShapeFeaset
	filename string //  filename
	data.Featuresets
	params data.ConnParams
}

// 打开一个shape文件，params["filename"] = "c:/data/a.shp"
func (this *ShapeStore) Open(params data.ConnParams) (bool, error) {
	this.params = params
	this.filename = params["filename"].(string)
	this.feaset = new(ShapeFeaset)
	this.feaset.store = this
	this.feaset.Name = base.GetTitle(this.filename)
	this.feaset.filename = this.filename
	this.Feasets = make([]data.Featureset, 1)
	this.Feasets[0] = this.feaset
	return true, nil
}

func (this *ShapeStore) GetConnParams() data.ConnParams {
	this.params["type"] = string(this.GetType())
	return this.params
}

// 得到存储类型
func (this *ShapeStore) GetType() data.StoreType {
	return data.StoreShape
}

// 关闭，释放资源
func (this *ShapeStore) Close() {
	this.feaset.Close()
}

// todo
func (this *ShapeStore) CreateFeaset(info data.FeasetInfo) data.Featureset {
	return nil
}

// todo
func (this *ShapeStore) DeleteFeaset(name string) bool {
	feaset, i := this.GetFeasetByName(name)
	if feaset != nil {
		feaset.Close()
		//todo 这里应该还要删除shp/dbf/shx/prj等磁盘文件
		this.Feasets = append(this.Feasets[:i], this.Feasets[i+1:]...)
		return true
	}
	return false
}
