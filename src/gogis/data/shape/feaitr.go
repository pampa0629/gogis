// shape存储库，硬盘模式
package shape

import (
	"gogis/base"
	"gogis/data"
	"gogis/geometry"
	"os"
	"sort"

	"github.com/LindsayBradford/go-dbf/godbf"
)

// shape读取迭代器
type ShapeFeaItr struct {
	ids        []int64      // id数组
	feaset     *ShapeFeaset // 数据集指针
	fields     []string     // 字段名，nil则为所有字段
	idss       [][]int64    // for batch fetch
	squery     data.SpatailQuery
	where      string
	countPerGo int
}

func (this *ShapeFeaItr) Count() int64 {
	return int64(len(this.ids))
}

func (this *ShapeFeaItr) Close() {
	this.ids = this.ids[:0]
	this.fields = this.fields[:0]
	this.feaset = nil
}

// 为了批量读取做准备，返回批量的次数
func (this *ShapeFeaItr) BeforeNext(countPerGo int) int {
	// 给ids排序，以便后面的连续读取
	// tr := base.NewTimeRecorder()
	if !this.feaset.isCache {
		sort.Sort(base.Int64s(this.ids))
		// tr.Output("sort ids")
	}

	goCount := len(this.ids)/countPerGo + 1
	// 这里假设每个code中所包含的对象，是大体平均分布的
	this.idss = base.SplitSlice64(this.ids, goCount)
	// tr.Output("split ids")
	this.countPerGo = countPerGo
	return goCount
}

// 得到需要的字段的 索引和类型
func (this *ShapeFeaItr) getFieldInfos() ([]string, []int, []godbf.DbaseDataType) {
	fields := this.fields
	if fields == nil { // nil表示全部字段都要
		fields = make([]string, len(this.feaset.FieldInfos))
		for i, v := range this.feaset.FieldInfos {
			fields[i] = v.Name
		}
	}

	var fieldCount int
	dbf := this.feaset.shape.table
	if dbf != nil {
		fieldInfos := dbf.Fields()
		fieldCount = len(fields)
		findexs := make([]int, fieldCount)
		ftypes := make([]godbf.DbaseDataType, fieldCount)
		for i, v := range fields {
			for j, finfo := range fieldInfos {
				if v == finfo.Name() {
					findexs[i] = j
					ftypes[i] = finfo.FieldType()
					break
				}
			}
		}
		return fields, findexs, ftypes
	}
	return nil, nil, nil
}

func (this *ShapeFeaItr) getGeosFromFile(ids []int64) []geometry.Geometry {
	count := len(ids)
	geos := make([]geometry.Geometry, count)

	if count > 0 {
		f, _ := os.Open(this.feaset.filename)
		defer f.Close()

		curPos := 0 // 当前位置
		for {
			// 这里要注意：为了保证至少读取一个对象，故而起始值为1；后续的判断要以此为基础开展计算
			batchCount := 1 // 连续的id的数量
			for curPos+batchCount+1 <= count {
				// 只有id连续，才能调用shape的Batch
				if ids[curPos+batchCount-1]+1 == ids[curPos+batchCount] {
					batchCount++
				} else {
					break
				}
			}
			this.feaset.shape.BatchLoad(f, int(ids[curPos]), batchCount, geos[curPos:], nil)
			curPos += batchCount
			if curPos >= count {
				break
			}
		}
	}
	return geos
}

func (this *ShapeFeaItr) getGeosFromCache(ids []int64) []geometry.Geometry {
	count := len(ids)
	geos := make([]geometry.Geometry, count)
	for i, v := range ids {
		geos[i] = this.feaset.geos[v]
	}
	return geos
}

// 批量获取，协程安全
func (this *ShapeFeaItr) BatchNext(batchNo int) ([]data.Feature, bool) {
	if int(batchNo) < len(this.idss) {
		ids := this.idss[batchNo]
		count := len(ids)

		var geos []geometry.Geometry
		if this.feaset.isCache {
			geos = this.getGeosFromCache(ids)
		} else {
			geos = this.getGeosFromFile(ids)
		}

		features := make([]data.Feature, 0, count)
		fields, findexs, ftypes := this.getFieldInfos()
		fieldCount := len(fields)

		var comps data.FieldComps
		comps.Parse(this.where, this.feaset.FieldInfos)
		for _, v := range geos {
			id := int(v.GetID())
			matchAtts := true
			atts := make(map[string]interface{}, fieldCount)
			if this.feaset.shape.table != nil {
				for j := 0; j < fieldCount; j++ {
					atts[fields[j]] = dbfString2Value(this.feaset.shape.table.FieldValue(id, findexs[j]), ftypes[j])
				}
				matchAtts = comps.Match(atts)
			}

			// 空间关系和 where都ok
			if this.squery.Match(v) && matchAtts {
				features = append(features, data.Feature{Geo: v, Atts: atts})
			}
		}

		return features, true
	}
	return nil, false
}

// 批量读取支持go协程安全
// func (this *ShapeFeaItr) BatchNext2(pos int64, count int) (features []Feature, newPos int64, result bool) {
// 	len := len(this.ids)
// 	if int(pos) < len {
// 		oldpos := int(pos)
// 		if count+int(pos) > len {
// 			count = len - int(pos)
// 		}
// 		pos += int64(count)
// 		features = make([]Feature, count)

// 		f, _ := os.Open(this.feaset.filename)
// 		defer f.Close()

// 		curPos := int(oldpos) // 当前位置
// 		for {
// 			// 这里要注意：为了保证至少读取一个对象，故而起始值为1；后续的判断要以此为基础开展计算
// 			batchCount := 1 // 连续的id的数量
// 			for curPos+batchCount+1 <= int(pos) {
// 				// 只有id连续，才能调用shape的Batch
// 				if this.ids[curPos+batchCount] == this.ids[curPos+batchCount-1]+1 {
// 					batchCount++
// 				} else {
// 					break
// 				}
// 			}
// 			this.feaset.shape.BatchLoad(f, int(this.ids[curPos]), batchCount, features[curPos-oldpos:], nil)
// 			curPos += batchCount
// 			if curPos >= int(pos) {
// 				break
// 			}
// 		}
// 		result = true
// 	}
// 	newPos = pos
// 	return
// }
