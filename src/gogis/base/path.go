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
	filenameall := path.Base(fullname)
	filesuffix := path.Ext(fullname)
	fileprefix := filenameall[0 : len(filenameall)-len(filesuffix)]
	return fileprefix
}

// 判断文件是否存在
func FileIsExist(filename string) bool {
	_, err := os.Stat(filename) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 两个绝对路径，得到path2相对于path1的路径
// 即基于path1，通过result可以得到path2
func GetRelativePath(path1, path2 string) string {
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

// 通过绝对路径+相对路径，得到绝对路径
// 例如：c:/temp/a.b + ./c.d --> c:/temp/c.d
func GetAbsolutePath(p, r string) string {
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
