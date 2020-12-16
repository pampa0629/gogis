// 路径相关的功能

package base

import (
	"fmt"
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

func GetRelativePath(bathpath, targetpath string) string {
	bathpath = filepath.Dir(bathpath) // 先得到路径
	relpath, err := filepath.Rel(bathpath, targetpath)
	if err != nil {
		fmt.Println("GetRelativePath err:", err)
	}
	return relpath
}

// 两个绝对路径，得到path2相对于path1的路径
// 即基于path1，通过result可以得到path2
func GetRelativePath2(path1, path2 string) string {
	// if path1 == "" || path2 == "" {
	// 	return "", errors.New("path cannot be empty")
	// }
	arr1 := strings.Split(path1[1:], "/")
	arr2 := strings.Split(path2[1:], "/")
	depth := 0
	for i := 0; i < len(arr1) && i < len(arr2); i++ {
		if arr1[i] == arr2[i] {
			depth++
		} else {
			break
		}
	}
	prefix := ""
	if len(arr1)-depth-1 <= 0 {
		prefix = "./"
	} else {
		for i := len(arr1) - depth - 1; i > 0; i-- {
			prefix += "../"
		}
	}
	fmt.Println(depth)
	if len(arr2)-depth > 0 {
		prefix += strings.Join(arr2[depth:], "/")
	}
	return prefix
}

func GetAbsolutePath(basepath, relpath string) string {
	basepath = filepath.Dir(basepath)
	abspath := filepath.Clean(filepath.Join(basepath, relpath))
	abspath = filepath.ToSlash(abspath)
	return abspath
}

// 通过绝对路径+相对路径，得到绝对路径
// 例如：c:/temp/a.b + ./c.d --> c:/temp/c.d
func GetAbsolutePath2(p, r string) string {
	p = toLinux(p)
	if "" == r || "." == r {
		return toLinux(p)
	}
	var linuxPath = toLinux(r)
	var paths = strings.Split(linuxPath, "/")

	var rp string
	if 0 < len(paths) {
		switch paths[0] {
		case ".": // . 需要去掉 p 中最后多余的部分
			// fallthrough
			p := path.Dir(p)
			rp = toLinux(p)
		case "":
			fallthrough
		case "..": // .. 直接保留即可
			rp = toLinux(p)
		}
	}
	var realPaths []string
	if "" != rp {
		realPaths = strings.Split(rp, "/")
	}
	if 0 < len(paths) {
		realPaths = append(realPaths, paths...)
	}

	result := path.Join(realPaths...)
	// 防止 最前面的 / 丢失
	if len(realPaths) > 0 && realPaths[0] == "" {
		result = "/" + result
	}

	return result
}
