// 数据转化类

package data

import (
	"fmt"
	"sync"
)

type Converter struct {
}

const CONVERT_GEO_COUNT_PERCOUNT = 1000
const CONVERT_BATCH_COUNT = 200

// 计算得到合理的批量数字，主要是  hbase 支持的写入端不能太多
func getGoodBatch(itr FeatureIterator, count int) (goCount int) {
	for {
		goCount = itr.PrepareBatch(count)
		if goCount > CONVERT_BATCH_COUNT {
			count *= 2 // 每次翻番，以减少go count
		} else {
			break
		}
	}
	return goCount
}

func (this *Converter) Convert(fromParams ConnParams, feasetName string, toParams ConnParams) {
	fromType := fromParams["type"]
	toType := toParams["type"]
	fromStore := NewDatastore(StoreType(fromType))
	toStore := NewDatastore(StoreType(toType))

	fromStore.Open(fromParams)
	toStore.Open(toParams)
	fromFeaset, _ := fromStore.GetFeasetByName(feasetName)
	fromFeaset.Open()

	toFeaset := toStore.(*HbaseStore).CreateFeaset(feasetName, fromFeaset.GetBounds(), fromFeaset.GetGeoType())
	// toFeaset.Open()
	fromItr := fromFeaset.QueryByBounds(fromFeaset.GetBounds())
	// forCount := fromItr.PrepareBatch(CONVERT_GEO_COUNT_PERCOUNT)
	forCount := getGoodBatch(fromItr, CONVERT_GEO_COUNT_PERCOUNT)

	outs := make([]int, forCount)

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < int(forCount); i++ {
		wg.Add(1)
		go this.batchConvert(fromItr, i, toFeaset, outs, wg)
	}
	wg.Wait()
	toFeaset.(*HbaseFeaset).EndWrite()

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

func (this *Converter) batchConvert(fromItr FeatureIterator, batchNo int, toFeaset Featureset, out []int, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	feas, ok := fromItr.BatchNext(batchNo)
	if ok {
		toFeaset.(*HbaseFeaset).BatchWrite(feas)
	}
	out[batchNo] = len(feas)
	// fmt.Println("convert batch no:", batchNo, "count:", len(feas))
}
