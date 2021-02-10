package hbase

import (
	"context"
	"encoding/binary"
	"fmt"
	"gogis/base"
	"gogis/data"
	"gogis/geometry"
	"io"

	"github.com/tsuna/gohbase/hrpc"
)

type HbaseFeaItr struct {
	feaset *HbaseFeaset
	bbox   base.Rect2D
	codes  []int32
	count  int64 // 对象数，猜测的

	countPerGo int       // 每一个批次的对象数量
	codess     [][]int32 // 每个批次所对应的index codes
}

func (this *HbaseFeaItr) Count() int64 {
	return this.count
}

// todo
// Next() (Feature, bool)

// 为调用批量读取做准备，调用 BatchNext 之前必须调用 本函数
// objCount 为每个批次拟获取对象的数量，不保证精确
func (this *HbaseFeaItr) BeforeNext(objCount int) int {
	goCount := int(this.count)/objCount + 1
	// 这里假设每个code中所包含的对象，是大体平均分布的
	this.codess = base.SplitSlice32(this.codes, goCount)
	if len(this.codess) != goCount {
		fmt.Println("PrepareBatch error! len of codes:", this.codes, "go count:", goCount)
	}
	// fmt.Println("codes:", this.codes)
	// fmt.Println("codess:", this.codess)
	this.countPerGo = objCount
	return goCount
}

// 把数组是否连续进行分割
func buildNextSlices(in []int32) (outs [][]int32) {
	var out []int32
	for pos := 0; pos < len(in); pos++ {
		out = append(out, in[pos])
		if pos == len(in)-1 {
			// 已经到最后一个，加自己
			outs = append(outs, out)
		} else {
			// 自己不是最后一个，则看下一个是否连续
			if in[pos+1] != in[pos]+1 {
				// 不连续就中断
				outs = append(outs, out)
				out = make([]int32, 0) // 必须重新make，防止冲突
			}
		}
	}
	return
}

// func buildNextSlices2(in []int) (outs [][]int) {
// 	outs = make([][]int, len(in))
// 	for i, v := range in {
// 		outs[i] = make([]int, 1)
// 		outs[i][0] = v
// 	}
// 	return
// }

// 批量读取，支持go协程安全；调用前，务必调用 PrepareBatch
// batchNo 为批量的序号
// 只要读取到一个数据，达不到count的要求，也返回true
// func (this *HbaseFeaItr) BatchNext2(batchNo int) (feas []data.Feature, result bool) {
// 	if batchNo < len(this.codess) {
// 		result = true // 只要no不越界，就返回 true
// 		codes := this.codess[batchNo]
// 		feas = make([]data.data.Feature, 0, this.countPerGo)

// 		codess := buildNextSlices(codes)
// 		// fmt.Println("BatchNext codess:", codess)

// 		var wg *sync.WaitGroup = new(sync.WaitGroup)
// 		feass := make([][]data.Feature, len(codess))
// 		for i, v := range codess {
// 			n := len(v) // 这个 v 中的id是连续的
// 			start := getRowkey(int32(v[0]), 0)
// 			end := getRowkey(int32(v[n-1]+1), 0)

// 			wg.Add(1)
// 			go this.batchNext(feass, i, start, end, wg, v[0], v[n-1])
// 		}
// 		wg.Wait()

// 		for _, v := range feass {
// 			feas = append(feas, v...)
// 		}
// 		// fmt.Println("feas count:", len(feas))
// 	}

// 	return
// }

// func (this *HbaseFeaItr) batchNext(feass [][]Feature, num int, start, end []byte, wg *sync.WaitGroup, code1, code2 int32) {
// 	defer wg.Done()

// 	// client := gohbase.NewClient(this.feaset.store.address)
// 	client := this.feaset.store.openClient()
// 	scanRequest, _ := hrpc.NewScanRange(context.Background(), []byte(this.feaset.name), start, end)
// 	scan := client.Scan(scanRequest)
// 	// count := 0
// 	for {
// 		var fea Feature
// 		getRsp, err := scan.Next()
// 		if err == io.EOF || getRsp == nil {
// 			break
// 		}
// 		if err != nil {
// 			fmt.Println("scan next error:", err)
// 		}
// 		for _, v := range getRsp.Cells {
// 			if string(v.Family) == "geo" && string(v.Qualifier) == "geom" {
// 				// fmt.Println("rowkey:", v.Row)
// 				fea.Geo = geometry.CreateGeo(this.feaset.geotype)
// 				fea.Geo.From(v.Value, geometry.WKB)
// 				// 用big，和构建row key保持一致
// 				id := int64(binary.BigEndian.Uint64(v.Row[4:]))
// 				fea.Geo.SetID(id)
// 				break
// 			}
// 		}
// 		// bbox相交，才取出去
// 		if this.bbox.IsIntersects(fea.Geo.GetBounds()) {
// 			// count++
// 			feass[num] = append(feass[num], fea)
// 		}
// 	}
// 	// fmt.Println("code1:", code1, "code2:", code2, "geo count:", len(feass[num]), "start:", start, "end:", end)
// 	// client.Close()
// 	this.feaset.store.closeClient(client)
// }

func (this *HbaseFeaItr) BatchNext(batchNo int) (feas []data.Feature, result bool) {
	if batchNo < len(this.codess) {
		result = true // 只要no不越界，就返回 true
		codes := this.codess[batchNo]
		feas = make([]data.Feature, 0, this.countPerGo)

		codess := buildNextSlices(codes)

		for _, v := range codess {
			n := len(v) // 这个 v 中的id是连续的
			// client := gohbase.NewClient(this.feaset.store.address)
			client := this.feaset.store.openClient()
			start := getRowkey(int32(v[0]), 0)
			end := getRowkey(int32(v[n-1]+1), 0)
			scanRequest, _ := hrpc.NewScanRange(context.Background(), []byte(this.feaset.Name), start, end)
			scan := client.Scan(scanRequest)
			count := 0
			// var rowkey []byte
			for {
				var fea data.Feature
				getRsp, err := scan.Next()
				if err == io.EOF || getRsp == nil {
					break
				}
				if err != nil {
					fmt.Println("scan next error:", err)
				}
				for _, vv := range getRsp.Cells {
					if string(vv.Family) == "geo" && string(vv.Qualifier) == "geom" {
						fea.Geo = geometry.CreateGeo(this.feaset.GeoType)
						fea.Geo.From(vv.Value, geometry.WKB)
						// 用big，和构建row key保持一致
						id := int64(binary.BigEndian.Uint64(vv.Row[4:]))
						fea.Geo.SetID(id)
						count++
						break
					}
				}
				// bbox相交，才取出去
				if this.bbox.IsIntersects(fea.Geo.GetBounds()) {
					feas = append(feas, fea)
				}
			}
			// client.Close()
			this.feaset.store.closeClient(client)
		}
	}

	return
}

func (this *HbaseFeaItr) Close() {
	this.codes = this.codes[:0]
	this.codess = this.codess[:0]
}
