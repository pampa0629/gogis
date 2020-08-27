package base

import (
	"io/ioutil"
	"os"
	"path"
)

// 删除空目录
func DeleteEmptyDir(path string) {
	dir, _ := ioutil.ReadDir(path)
	if len(dir) == 0 {
		os.RemoveAll(path)
	}
}

// 得到文件名中的 title部分；输入：C:/temp/JBNTBHTB.shp ，返回 JBNTBHTB
func GetTile(fullname string) string {
	filenameall := path.Base(fullname)
	filesuffix := path.Ext(fullname)
	fileprefix := filenameall[0 : len(filenameall)-len(filesuffix)]
	return fileprefix
}
