// 路径相关的功能

package base

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// 获得程序运行路径
func RootPath() string {
	var fp, _ = filepath.Abs(path.Dir(os.Args[0]))
	return fp
}

// 统一转化为linux风格的路径
func toLinux(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// 删除空目录
func DeleteEmptyDir(path string) {
	dir, _ := ioutil.ReadDir(path)
	if len(dir) == 0 {
		os.RemoveAll(path)
	}
}

// 得到文件名中的 title部分；输入：C:/temp/JBNTBHTB.shp ，返回 JBNTBHTB
func GetTitle(fullname string) string {
	fullname = toLinux(fullname)
	filename := path.Base(fullname)
	extname := path.Ext(filename)
	titlename := filename[0 : len(filename)-len(extname)]
	return titlename
}

// 得到文件名中的 ext部分；输入：C:/temp/JBNTBHTB.shp ，返回 shp；
// 注意：没有点号"."
func GetExt(fullname string) string {
	return strings.TrimLeft(path.Ext(fullname), ".")
}

// 改变扩展名；输入 c:/temp/a.bcd和efg, 输出 c:/temp/a.efg
func ChangeExt(filename, ext string) string {
	return strings.TrimSuffix(filename, path.Ext(filename)) + "." + ext
}

// 判断文件或文件夹是否存在
func IsExist(filepath string) bool {
	_, err := os.Stat(filepath) //os.Stat获取文件信息
	return err == nil || os.IsExist(err)
}

// 判断文件是否存在；若存在的是文件夹，也返回false
func FileIsExist(filename string) bool {
	fi, err := os.Stat(filename) //os.Stat获取文件信息
	return (err == nil || os.IsExist(err)) && !fi.IsDir()
}

// 判断文件夹是否存在；若存在的是文件，也返回false
func DirIsExist(pathname string) bool {
	fi, err := os.Stat(pathname) //os.Stat获取文件信息
	return (err == nil || os.IsExist(err)) && fi.IsDir()
}

// 两个绝对路径，得到targetpath相对于bathpath的路径;
// 即基于bathpath，通过result可以得到targetpath；
// 如：GetRelativePath("c:/temp/", "c:/abc.def") == "../abc.def"
func GetRelativePath(bathpath, targetpath string) string {
	bathpath = filepath.Dir(bathpath) // 先得到路径
	relpath, err := filepath.Rel(bathpath, targetpath)
	if err != nil { // 有错误，就直接返回
		return targetpath
	}
	return relpath
}

// 通过绝对路径+相对路径，得到绝对路径
// 例如：c:/temp/a.b + ./c.d --> c:/temp/c.d
func GetAbsolutePath(basepath, relpath string) string {
	basepath = filepath.Dir(basepath)
	abspath := filepath.Clean(filepath.Join(basepath, relpath))
	abspath = filepath.ToSlash(abspath)
	return abspath
}
