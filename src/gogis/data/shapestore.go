// shape存储库，硬盘模式

package data

import (
	"errors"
	"fmt"
	"gogis/base"
	"gogis/index"
	"os"
	"sort"
	"strconv"
	"strings"
)

// 快捷方法，打开一个shape文件，得到特征集对象
func OpenShape(filename string) Featureset {
	// 默认用内存模式
	shp := new(ShpmemStore)
	params := NewConnParams()
	params["filename"] = filename
	shp.Open(params)
	feaset, _ := shp.GetFeasetByNum(0)
	return feaset
}

// todo 未来还要考虑实现打开一个文件夹
type ShapeStore struct {
	feaset *ShapeFeaset
	name   string //  filename
}

// 打开一个shape文件，params["filename"] = "c:/data/a.shp"
func (this *ShapeStore) Open(params ConnParams) (bool, error) {
	this.feaset = new(ShapeFeaset)
	this.feaset.store = this
	this.name = params["filename"]
	return this.feaset.Open(this.name)
}

func (this *ShapeStore) GetConnParams() ConnParams {
	params := NewConnParams()
	params["filename"] = this.name
	params["type"] = string(this.GetType())
	return params
}

// 得到存储类型
func (this *ShapeStore) GetType() StoreType {
	return StoreShape
}

func (this *ShapeStore) GetFeasetByNum(num int) (Featureset, error) {
	if num == 0 {
		return this.feaset, nil
	}
	return nil, errors.New(strconv.Itoa(num) + " beyond the num of feature sets")
}

func (this *ShapeStore) GetFeasetByName(name string) (Featureset, error) {
	if this.feaset.name == name {
		return this.feaset, nil
	}
	return nil, errors.New("feature set: " + name + " cannot find")
}

func (this *ShapeStore) FeaturesetNames() []string {
	names := make([]string, 1)
	names[0] = this.feaset.name
	return names
}

// 关闭，释放资源
func (this *ShapeStore) Close() {
	// this.MemoryStore.Close()
	this.feaset.Close()
}

// 硬盘读写模式的shape数据集
type ShapeFeaset struct {
	// MemFeaset

	name string
	bbox base.Rect2D
	// shape文件
	shape *ShapeFile
	// 空间索引
	index index.SpatialIndex

	store *ShapeStore
}

// 打开shape文件
func (this *ShapeFeaset) Open(filename string) (bool, error) {
	tr := base.NewTimeRecorder()
	this.name = base.GetTitle(filename)

	this.shape = new(ShapeFile)
	res := this.shape.Open(filename)
	this.bbox = base.NewRect2D(this.shape.Xmin, this.shape.Ymin, this.shape.Xmax, this.shape.Ymax)

	// this.loadShape(shape)
	// shape.Close()

	tr.Output("open shape file: " + filename + ",")

	//  处理空间索引文件
	this.loadSpatialIndex()

	//  todo 矢量金字塔
	// this.BuildPyramids()

	return res, nil
}

func (this *ShapeFeaset) Close() {
	this.shape.Close()
	this.store = nil
}

func (this *ShapeFeaset) GetName() string {
	return this.name
}

// 创建或者加载空间索引文件
func (this *ShapeFeaset) loadSpatialIndex() {
	indexName := strings.TrimSuffix(this.store.name, ".shp") + "." + base.EXT_SPATIAL_INDEX_FILE
	if base.FileIsExist(indexName) {
		this.index = index.LoadGix(indexName)
	} else {
		indexType := index.TypeQTreeIndex // 这里确定索引类型 TypeQTreeIndex TypeRTreeIndex TypeGridIndex
		spatialIndex := this.BuildSpatialIndex(indexType)
		index.SaveGix(indexName, spatialIndex)
	}
}

func (this *ShapeFeaset) BuildSpatialIndex(indexType index.SpatialIndexType) index.SpatialIndex {
	if this.index == nil {
		tr := base.NewTimeRecorder()

		this.index = index.NewSpatialIndex(indexType)
		this.index.Init(this.bbox, this.Count())
		bboxes, ids := this.shape.LoadBboxIds()
		for i, v := range bboxes {
			this.index.AddOne(v, int64(ids[i]))
		}
		check := this.index.Check()
		fmt.Println("check building spatial index, result is:", check)

		tr.Output("build spatial index")
		return this.index
	}
	return nil
}

func (this *ShapeFeaset) GetStore() Datastore {
	return this.store
}

func (this *ShapeFeaset) Count() int64 {
	return int64(this.shape.recordNum)
}

func (this *ShapeFeaset) GetBounds() base.Rect2D {
	return this.bbox
}

func (this *ShapeFeaset) GetFieldInfos() []FieldInfo {
	return this.shape.GetFieldInfos()
}

// todo
func (this *ShapeFeaset) Query(bbox base.Rect2D, def QueryDef) FeatureIterator {
	return nil
}

func (this *ShapeFeaset) QueryByBounds(bbox base.Rect2D) FeatureIterator {
	feaItr := new(ShapeFeaItr)
	feaItr.feaset = this
	feaItr.ids = this.index.Query(bbox)
	// 给ids排序，以便后面的连续读取
	sort.Sort(base.Int64s(feaItr.ids))
	return feaItr
}

// todo
func (this *ShapeFeaset) QueryByDef(def QueryDef) FeatureIterator {
	return nil
}

// shape读取迭代器
type ShapeFeaItr struct {
	ids        []int64      // id数组
	feaset     *ShapeFeaset // 数据集指针
	fields     []string     // 字段名，空则为所有字段
	idss       [][]int64    // for batch fetch
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

// todo
func (this *ShapeFeaItr) Next() (Feature, bool) {
	// if this.pos < len(this.ids) {
	// 	oldpos := this.pos
	// 	this.pos++
	// 	id := this.ids[oldpos]
	// 	id = 0 // todo
	// 	// todo
	// 	return *new(Feature), false
	// 	// return this.feaset.shape.LoadOne(this.file, int(id)), true
	// } else {
	// 	return *new(Feature), false
	// }
	return *new(Feature), false
}

// 为了批量读取做准备，返回批量的次数
func (this *ShapeFeaItr) PrepareBatch(countPerGo int) int {
	goCount := len(this.ids)/countPerGo + 1
	// 这里假设每个code中所包含的对象，是大体平均分布的
	this.idss = splitSlice64(this.ids, goCount)
	this.countPerGo = countPerGo
	return goCount
}

// 批量获取，协程安全
func (this *ShapeFeaItr) BatchNext(batchNo int) (features []Feature, result bool) {

	if int(batchNo) < len(this.idss) {

		count := len(this.idss[batchNo])
		ids := this.idss[batchNo]
		features = make([]Feature, count)

		f, _ := os.Open(this.feaset.store.name)
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
			this.feaset.shape.BatchLoad(f, int(ids[curPos]), batchCount, features[curPos:], nil)
			curPos += batchCount
			if curPos >= count {
				break
			}
		}
		result = true
	}
	return
}

// 批量读取支持go协程安全
func (this *ShapeFeaItr) BatchNext2(pos int64, count int) (features []Feature, newPos int64, result bool) {
	// tr := base.NewTimeRecorder()

	len := len(this.ids)
	if int(pos) < len {
		oldpos := int(pos)
		if count+int(pos) > len {
			count = len - int(pos)
		}
		pos += int64(count)
		features = make([]Feature, count)

		f, _ := os.Open(this.feaset.store.name)
		defer f.Close()

		curPos := int(oldpos) // 当前位置
		for {
			// 这里要注意：为了保证至少读取一个对象，故而起始值为1；后续的判断要以此为基础开展计算
			batchCount := 1 // 连续的id的数量
			for curPos+batchCount+1 <= int(pos) {
				// 只有id连续，才能调用shape的Batch
				if this.ids[curPos+batchCount] == this.ids[curPos+batchCount-1]+1 {
					batchCount++
				} else {
					break
				}
			}
			this.feaset.shape.BatchLoad(f, int(this.ids[curPos]), batchCount, features[curPos-oldpos:], nil)
			curPos += batchCount
			if curPos >= int(pos) {
				break
			}
		}
		result = true
	}
	newPos = pos
	return
}
