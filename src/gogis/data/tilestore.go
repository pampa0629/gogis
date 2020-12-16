package data

import (
	"os"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// 瓦片类数据存储和读取接口
// todo 未来还需要增加：图片格式，时间（版本）之类的信息
type TileStore interface {
	// 打开，读写皆可，不存在时自动创建
	Open(path string, name string) (bool, error)
	// 根据层级等信息存储瓦片
	Put(level int, col int, row int, data []byte)
	// 根据层级等信息加载瓦片
	Get(level int, col int, row int) []byte
	Close()
}

// 文件方式存储
type FileTileStore struct {
	cachepath string // 缓存所在路径
}

// 打开，读写皆可，不存在时自动创建
func (this *FileTileStore) Open(path string, name string) (bool, error) {
	this.cachepath = path + "/" + name + "/"
	//  创建文件夹
	err := os.MkdirAll(this.cachepath, os.ModePerm)
	if err != nil {
		return false, err
	}
	return true, nil
}

// 根据层级等信息存储瓦片
func (this *FileTileStore) Put(level int, col int, row int, data []byte) {
	pathname := getPathName(this.cachepath, level, col)
	os.MkdirAll(pathname, os.ModePerm)
	filename := getFileName(pathname, row)
	f, _ := os.Create(filename)
	defer f.Close()
	if f != nil {
		f.Write(data)
	}
}

// 根据层级等信息加载瓦片
func (this *FileTileStore) Get(level int, col int, row int) (data []byte) {
	pathname := getPathName(this.cachepath, level, col)
	filename := getFileName(pathname, row)
	f, err := os.Open(filename)
	defer f.Close()

	if f != nil && err == nil {
		info, _ := os.Stat(filename)
		data = make([]byte, info.Size())
		f.Read(data)
	}
	return data
}

// 得到tile文件所在的目录名字
func getPathName(rootpath string, level int, col int) string {
	return rootpath + strconv.Itoa(level) + "/" + strconv.Itoa(col) + "/"
}

// 得到 tile文件名
func getFileName(pathname string, row int) string {
	// 当前默认是 png文件，后续还要支持其它图片格式
	filename := pathname + strconv.Itoa(row) + ".png"
	return filename
}

func (this *FileTileStore) Close() {
	// do nothing
}

// leveldb 存储
type LeveldbTileStore struct {
	db *leveldb.DB
}

const CACHE_FILE_SIZE = 1024 * 1024 * 100

// 打开，读写皆可，不存在时自动创建
func (this *LeveldbTileStore) Open(path string, name string) (bool, error) {
	dbfile := path + "/" + name
	db, err := leveldb.OpenFile(dbfile, &opt.Options{WriteBuffer: CACHE_FILE_SIZE})
	if db == nil || err != nil {
		return false, err
	}
	this.db = db
	return true, nil
}

// 根据层级等信息存储瓦片
func (this *LeveldbTileStore) Put(level int, col int, row int, data []byte) {
	key := getCacheKey(level, col, row)
	this.db.Put([]byte(key), data, nil)
}

// 根据层级等信息加载瓦片
func (this *LeveldbTileStore) Get(level int, col int, row int) []byte {
	key := getCacheKey(level, col, row)
	data, _ := this.db.Get([]byte(key), nil)
	return data
}

func (this *LeveldbTileStore) Close() {
	this.db.Close()
}

func getCacheKey(level int, col int, row int) string {
	return strconv.Itoa(level) + "/" + strconv.Itoa(col) + "/" + strconv.Itoa(row)
}
