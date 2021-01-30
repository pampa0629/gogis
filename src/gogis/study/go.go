package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"gogis/base"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	// pool "github.com/silenceper/poor"
)

func gomain() {
	testType()
	fmt.Println("DONE!")
}

type Point2Ds []base.Point2D

func (this *Point2Ds) Test() {
	fmt.Println("func (this *Point2Ds) Test() {")
}

func Test2(pnt Point2Ds) {
	fmt.Println("func (this *Point2Ds) Test() {", pnt)
}

func testType() {
	pnts := make([]base.Point2D, 3)
	ps := Point2Ds(pnts)
	ps.Test()
	Test2(pnts)
}

func testInterface() {
	var bbox base.Rect2D
	var inter interface{}
	inter = bbox
	name := reflect.TypeOf(inter)
	fmt.Println(name)
	// b1, o1 := inter.(base.Bounds)
	// fmt.Println(b1, o1)
	// var b2 base.Bounds = inter
	// b2, o2 := (&b1).(base.Bounds)
	// fmt.Println(b2)

}

func testFileCount() {
	// count := 5000                    // 最大支持并发
	// ch := make(chan struct{}, count) // 控制任务并发的chan
	// defer close(ch)
	var gomax base.GoMax
	gomax.Init(5000)

	path := "c:/temp/cache/test/"
	// var wg *sync.WaitGroup = new(sync.WaitGroup)
	for i := 0; i < 200; i++ {
		// ch <- struct{}{} // 作用类似于waitgroup.Add(1)
		gomax.Add()
		// wg.Add(1)
		go testGo(path, i, &gomax)
	}
	// wg.Wait()
	gomax.Wait()
	return
}

// func (this *TestCreate) testGo(path string, i int, wg *sync.WaitGroup) {
// func testGo(path string, i int, wg *sync.WaitGroup, gomax *GoMax) {
// func testGo(path string, i int, wg *sync.WaitGroup, ch chan struct{}) {
func testGo(path string, k int, gomax *base.GoMax) {
	if gomax != nil {
		defer gomax.Done()
	}

	for i := 0; i < 200; i++ {
		for j := 0; j < 1000; j++ {
			gomax.Add()
			go testGo2(path, k, i, j, gomax)
		}
	}
	// gomax.Done()
}

func testGo3(path string, i, j, k int, gomax *base.GoMax) {
	testGo2(path, k, i, j, gomax)

}

func testGo2(path string, i, j, k int, gomax *base.GoMax) {
	defer gomax.Done()
	filename := path + strconv.Itoa(i) + "-" + strconv.Itoa(j) + "-" + strconv.Itoa(k) + ".txt"
	f, _ := os.Create(filename)
	// gofile := this.openFile()
	// f := gofile.CreateFile(this.filename)
	f.WriteString(filename)
	defer f.Close()
	fmt.Println("i,j,k:", i, j, k)
	// <-ch // 执行完毕，释放资源
}

func testJson2() {
	str := `{"took":1,"timed_out":false,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0},"hits":{"total":{"value":2391,"relation":"eq"},"max_score":null,"hits":[]},"aggregations":{"viewport":{"bounds":{"top_left":{"lat":53.13844630960375,"lon":74.99301684089005},"bottom_right":{"lat":18.483061753213406,"lon":134.36185402795672}}}}}`
	fmt.Println(len(str), str)
	var maps1 map[string]interface{}
	// maps1 := make(map[string]interface{})
	json.Unmarshal([]byte(str), &maps1)
	fmt.Println([]byte(str))
	fmt.Println(maps1)
}

func testJson() {
	mappings := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"location": map[string]interface{}{
					"lat": 48.85999997612089,
					"lon": 2.3363889567553997,
				},
			},
		},
	}

	fmt.Println(mappings)
	// var buf bytes.Buffer
	// json.NewEncoder(&buf).Encode(mappings)
	data, _ := json.Marshal(mappings)
	fmt.Println(string(data))

	// json.NewEncoder(&buf).Encode(mappings)

	// str := "aggregations:{viewport:{bounds:{top_left:{lat:53.13844630960375,lon:74.99301684089005},bottom_right:{lat:18.483061753213406,lon:134.36185402795672}}}}"
	// var maps map[string]interface{}
	var maps1 map[string]interface{}
	json.Unmarshal(data, &maps1)
	fmt.Println(string(data))
	// json.Unmarshal([]byte(str), &maps)
	// json.Unmarshal(buf.Bytes(), &maps1)
	// json.NewDecoder(&buf).Decode(&maps1)
}

func testFilePath() {

	//返回所给路径的绝对路径
	path, _ := filepath.Abs("./1.txt")
	fmt.Println(path)

	//返回路径最后一个元素
	fmt.Println(filepath.Base("./1.txt"))
	//如果路径为空字符串，返回.
	fmt.Println(filepath.Base(""))
	//如果路径只有斜线，返回/
	fmt.Println(filepath.Base("///"))

	//返回等价的最短路径
	//1.用一个斜线替换多个斜线
	//2.清除当前路径.
	//3.清除内部的..和他前面的元素
	//4.以/..开头的，变成/
	fmt.Println(filepath.Clean("C:/a/b/../c"))
	fmt.Println(filepath.Clean("./1.txt"))

	//返回路径最后一个元素的目录
	//路径为空则返回.
	fmt.Println(filepath.Dir("./a/b/c"))
	fmt.Println(filepath.Dir("C:/a/b/c"))

	//返回链接文件的实际路径
	path2, _ := filepath.EvalSymlinks("1.lnk")
	fmt.Println(path2)

	//返回路径中的扩展名
	//如果没有点，返回空
	fmt.Println(filepath.Ext("./a/b/c/d.jpg"))

	//将路径中的/替换为路径分隔符
	fmt.Println(filepath.FromSlash("./a/b/c"))

	//返回所有匹配的文件
	match, _ := filepath.Glob("./*.go")
	fmt.Println(match)

	//判断路径是不是绝对路径
	fmt.Println(filepath.IsAbs("./a/b/c"))
	fmt.Println(filepath.IsAbs("C:/a/b/c"))

	//连接路径，返回已经clean过的路径
	fmt.Println(filepath.Join("C:/a", "/b", "/c"))

	//匹配文件名，完全匹配则返回true
	fmt.Println(filepath.Match("*", "a"))
	fmt.Println(filepath.Match("*", "C:/a/b/c"))
	fmt.Println(filepath.Match("\\b", "b"))

	//返回以basepath为基准的相对路径
	path3, _ := filepath.Rel("C:/a/b", "C:/a/b/c/d/../e")
	fmt.Println(path3)

	//将路径使用路径列表分隔符分开，见os.PathListSeparator
	//linux下默认为:，windows下为;
	fmt.Println(filepath.SplitList("C:/windows;C:/windows/system"))

	//分割路径中的目录与文件
	dir, file := filepath.Split("C:/a/b/c/d.jpg")
	fmt.Println(dir, file)

	//将路径分隔符使用/替换
	fmt.Println(filepath.ToSlash("C:/a/b"))

	//返回分区名
	fmt.Println(filepath.VolumeName("C:/a/b/c"))

}

func testPath() {

	p1 := "c:/temp/zengzm/abc.ext"
	p2 := "c:/temp/zengzm/def.ext"
	{
		b11 := filepath.Base(p1)
		fmt.Println(b11)
		b12 := filepath.Base(p2)
		fmt.Println(b12)
		b21 := path.Base(p1)
		fmt.Println(b21)
		b22 := path.Base(p2)
		fmt.Println(b22)
	}

	//返回路径的最后一个元素
	fmt.Println(path.Base("./a/b/c"))
	//如果路径为空字符串，返回.
	fmt.Println(path.Base(""))
	//如果路径只有斜线，返回/
	fmt.Println(path.Base("///"))

	//返回等价的最短路径
	//1.用一个斜线替换多个斜线
	//2.清除当前路径.
	//3.清除内部的..和他前面的元素
	//4.以/..开头的，变成/
	fmt.Println(path.Clean("./a/b/../"))

	//返回路径最后一个元素的目录
	//路径为空则返回.
	fmt.Println(path.Dir("./a/b/c"))

	//返回路径中的扩展名
	//如果没有点，返回空
	fmt.Println(path.Ext("./a/b/c/d.jpg"))

	//判断路径是不是绝对路径
	fmt.Println(path.IsAbs("./a/b/c"))
	fmt.Println(path.IsAbs("/a/b/c"))

	//连接路径，返回已经clean过的路径
	fmt.Println(path.Join("./a", "b/c", "../d/"))

	//匹配文件名，完全匹配则返回true
	fmt.Println(path.Match("*", "a"))
	fmt.Println(path.Match("*", "a/b/c"))
	fmt.Println(path.Match("\\b", "b"))

	//分割路径中的目录与文件
	fmt.Println(path.Split("./a/b/c/d.jpg"))

}

func returnTow() (int, int) {
	return 0, 1
}

func testReturn() {
	a := 3
	a, b := returnTow()
	{
		a = 4
		a += 1
	}
	fmt.Println("a,b", a, b)
}

func testBuffer() {
	data := make([]byte, 0)
	buf := bytes.NewBuffer(data)
	buf.Write([]byte("abcde"))
	data = buf.Bytes()
	fmt.Println(data)
}

func testMath() {
	n := 10 / 100
	fmt.Println("n:", n)
}

func testFileWR() {
	filename := "c:/temp/abc"
	{
		f, _ := os.Create(filename)

		// data := []byte{'a', 'b', 'c'}
		str := "abc"
		f.WriteString(str)

		// binary.Write(f, binary.LittleEndian, data)
		f.Close()
	}

	{
		gix, _ := os.Open(filename)

		// var gixMark [3]byte
		gixMark := make([]byte, 3)
		// gix.Read(gixMark)
		binary.Read(gix, binary.LittleEndian, gixMark)
		gix.Close()
	}

}
