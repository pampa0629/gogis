// 数据转化类

package data

import (
	"fmt"
	"gogis/base"
	"gogis/geometry"
	"runtime"
)

type Converter struct {
}

const CONVERT_GEO_COUNT_PERCOUNT = 10000
const CONVERT_BATCH_COUNT = 200

// 计算得到合理的批量数字，主要是  hbase 支持的写入端不能太多
func getGoodBatch(itr FeatureIterator, count int) (goCount int) {
	for {
		goCount = itr.BeforeNext(count)
		if goCount > CONVERT_BATCH_COUNT {
			count *= 2 // 每次翻番，以减少go count
		} else {
			break
		}
	}
	return goCount
}

func (this *Converter) Convert(fromParams ConnParams, feasetName string, toParams ConnParams) {
	fromType := fromParams["type"].(string)
	toType := toParams["type"].(string)
	fromStore := NewDatastore(StoreType(fromType))
	toStore := NewDatastore(StoreType(toType))

	fromStore.Open(fromParams)
	toStore.Open(toParams)
	fromFeaset, _ := fromStore.GetFeasetByName(feasetName)
	fromFeaset.Open()

	var info FeasetInfo
	info.Name = feasetName
	info.Bbox = fromFeaset.GetBounds()
	info.GeoType = fromFeaset.GetGeoType()
	info.Proj = fromFeaset.GetProjection()
	info.FieldInfos = fromFeaset.GetFieldInfos()
	toFeaset := toStore.CreateFeaset(info)
	fromItr := fromFeaset.Query(nil)
	toFeaset.BeforeWrite(fromItr.Count())
	// forCount := fromItr.PrepareBatch(CONVERT_GEO_COUNT_PERCOUNT)
	forCount := getGoodBatch(fromItr, CONVERT_GEO_COUNT_PERCOUNT)

	outs := make([]int, forCount)

	if forCount == 1 {
		this.batchConvert(fromItr, 0, toFeaset, outs, nil)
	} else if forCount > 1 {
		// var wg *sync.WaitGroup = new(sync.WaitGroup)
		var gm *base.GoMax = new(base.GoMax)
		params := toStore.GetConnParams()
		gm.Init(params["gowrite"].(int))
		for i := 0; i < int(forCount); i++ {
			// wg.Add(1)
			gm.Add()
			go this.batchConvert(fromItr, i, toFeaset, outs, gm)
		}
		gm.Wait()
	}
	toFeaset.EndWrite()

	count := 0
	for _, v := range outs {
		count += v
	}
	fmt.Println("total count:", count)

	fromItr.Close()
	fromFeaset.Close()
	toFeaset.Close()
	fromStore.Close()
	toStore.Close()
}

func (this *Converter) batchConvert(fromItr FeatureIterator, batchNo int, toFeaset Featureset, out []int, gm *base.GoMax) {
	if gm != nil {
		defer gm.Done()
	}

	feas, ok := fromItr.BatchNext(batchNo)
	if ok {
		toFeaset.BatchWrite(feas)
	}
	out[batchNo] = len(feas)
	// fmt.Println("convert batch no:", batchNo, "count:", len(feas))
}

// 线转点
func (this *Converter) Polyline2Point(feasetLine Featureset, toStore Datastore, name string) {
	if feasetLine.GetGeoType() != geometry.TGeoPolyline {
		return
	}
	var info FeasetInfo
	info.Bbox = feasetLine.GetBounds()
	info.FieldInfos = feasetLine.GetFieldInfos()
	info.GeoType = geometry.TGeoPoint
	info.Name = name
	info.Proj = feasetLine.GetProjection()
	feasetPoint := toStore.CreateFeaset(info)

	feait := feasetLine.Query(nil)

	objCount := feait.Count()
	feasetPoint.BeforeWrite(objCount)
	forCount := feait.BeforeNext(getObjCount(objCount))

	var gm *base.GoMax = new(base.GoMax)
	params := toStore.GetConnParams()
	gm.Init(params["gowrite"].(int))
	for i := 0; i < int(forCount); i++ {
		gm.Add()
		go this.goPolyline2Point(feait, i, feasetPoint, gm)
	}
	gm.Wait()
	feasetPoint.EndWrite()

	feait.Close()
	feasetLine.Close()
	feasetPoint.Close()
}

func (this *Converter) goPolyline2Point(fromItr FeatureIterator, batchNo int, toFeaset Featureset, gm *base.GoMax) {
	if gm != nil {
		defer gm.Done()
	}

	feas, ok := fromItr.BatchNext(batchNo)
	if ok {
		feaPnts := make([]Feature, 0, len(feas)*10)
		for _, v := range feas {
			subCount := v.Geo.SubCount()
			geo1 := v.Geo.(geometry.Geo1)
			for i := 0; i < subCount; i++ {
				line := geo1.GetSubLine(i)
				for _, pnt := range line {
					var feaPnt Feature
					var geoPnt geometry.GeoPoint
					geoPnt.Point2D = pnt
					feaPnt.Geo = &geoPnt
					feaPnts = append(feaPnts, feaPnt)
				}
			}
		}
		toFeaset.BatchWrite(feaPnts)
		feaPnts = feaPnts[:0]
	}
}

const ONE_GO_COUNT = 10000

func getObjCount(count int64) int {
	objCount := int(count) / runtime.NumCPU()
	objCount = base.IntMax(objCount, ONE_GO_COUNT)
	objCount = base.IntMin(objCount, ONE_GO_COUNT*20)
	return objCount
}
