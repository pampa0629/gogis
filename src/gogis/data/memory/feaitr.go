package memory

import (
	"gogis/data"
	"gogis/geometry"
)

type MemFeaItr struct {
	ids              []int64              // id数组
	feaset           *MemFeaset           // 数据集指针
	geoPyramid       *[]geometry.Geometry // 金字塔层的对象
	objCountPerBatch int                  // 每个批次要读取的对象数量
	fields           []string             // 字段名，空则为所有字段
	squery           data.SpatailQuery
}

func (this *MemFeaItr) Count() int64 {
	return int64(len(this.ids))
}

func (this *MemFeaItr) Close() {
	this.ids = this.ids[:0]
	this.fields = this.fields[:0]
	this.feaset = nil
}

// 为了批量读取做准备，返回批量的次数
func (this *MemFeaItr) BeforeNext(objCount int) int {
	goCount := len(this.ids)/objCount + 1
	this.objCountPerBatch = objCount
	return goCount
}

// 批量读取支持go协程安全
func (this *MemFeaItr) BatchNext(batchNo int) (feas []data.Feature, result bool) {
	remainCount := len(this.ids) - batchNo*this.objCountPerBatch
	if remainCount >= 0 {
		objCount := this.objCountPerBatch
		if remainCount < objCount {
			objCount = remainCount
		}
		start := batchNo * this.objCountPerBatch
		feas = this.getFeaturesByIds(this.ids[start : start+objCount])
		result = true
	}
	return
}

func (this *MemFeaItr) getFeaturesByIds(ids []int64) []data.Feature {
	feas := make([]data.Feature, 0, len(ids))
	for _, id := range ids {
		if fea, ok := this.getOneFeature(id); ok {
			feas = append(feas, fea)
		}
	}
	return feas
}

// 返回 false，说明这个不能要
func (this *MemFeaItr) getOneFeature(id int64) (fea data.Feature, res bool) {
	if this.geoPyramid != nil {
		fea.Geo = (*this.geoPyramid)[id].Clone()
	} else {
		fea.Geo = this.feaset.Features[id].Geo.Clone()
	}
	if this.squery.Match(fea.Geo) {
		this.setFeaAtts(fea, this.feaset.Features[id])
		res = true
	}

	return
}

// 根据需要，只取一部分字段值
func (this *MemFeaItr) setFeaAtts(out, fea data.Feature) {
	out.Atts = make(map[string]interface{})
	// nil表明：所有属性全要
	if this.fields == nil {
		for k, v := range fea.Atts {
			out.Atts[k] = v
		}
	} else {
		// 根据 fields 来设置属性
		for _, field := range this.fields {
			out.Atts[field] = fea.Atts[field]
		}
	}
}
