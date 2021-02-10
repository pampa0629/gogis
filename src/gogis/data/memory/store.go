package memory

import (
	"gogis/data"
	"runtime"
)

func init() {
	data.RegisterDatastore(data.StoreMemory, NewMemoryStore)
}

func NewMemoryStore() data.Datastore {
	return new(MemoryStore)
}

type MemoryStore struct {
	data.Featuresets
	data.ConnParams
}

// nothing to do
func (this *MemoryStore) Open(params data.ConnParams) (bool, error) {
	this.ConnParams = params
	this.ConnParams["cache"] = true
	params["gowrite"] = runtime.NumCPU()
	return true, nil
}

func (this *MemoryStore) GetConnParams() data.ConnParams {
	return this.ConnParams
}

// 得到存储类型
func (this *MemoryStore) GetType() data.StoreType {
	return data.StoreMemory
}

// 关闭，释放资源
func (this *MemoryStore) Close() {
	for _, feaset := range this.Feasets {
		feaset.Close()
	}
	this.Feasets = this.Feasets[:0]
}

func (this *MemoryStore) CreateFeaset(info data.FeasetInfo) data.Featureset {
	feaset := new(MemFeaset)
	feaset.FeasetInfo = info
	feaset.store = this
	this.Feasets = append(this.Feasets, feaset)
	return feaset
}

func (this *MemoryStore) DeleteFeaset(name string) bool {
	feaset, i := this.GetFeasetByName(name)
	if feaset != nil {
		feaset.Close()
		this.Feasets = append(this.Feasets[:i], this.Feasets[i+1:]...)
		return true
	}
	return false
}
