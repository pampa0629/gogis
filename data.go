package main

import (
	"gogis/base"
	"gogis/data"
	_ "gogis/data/memory"
)

func Convert(input, output, name string) {
	fromParams := data.NewConnParams()
	fromParams["filename"] = input
	fromParams["type"] = string(data.StoreShape)

	toParams := data.NewConnParams()
	toParams["filename"] = output
	toParams["type"] = string(data.StoreSqlite)

	var cvt data.Converter
	title := base.GetTitle(input)
	if len(name) == 0 {
		name = base.GetTitle(output)
	}
	cvt.Convert(fromParams, title, toParams, name)
}

func CreateIndex(datafile, indextype string) {
	if len(indextype) == 0 {
		indextype = "qtree" // 默认四叉树，比较均衡
	}
	if base.GetExt(datafile) == "shp" {
		store := data.NewDatastore(data.StoreShape)
		params := data.NewConnParams()
		params["filename"] = datafile
		params["index"] = indextype
		store.Open(params)
		feaset := store.GetFeasetByNum(0)
		if feaset != nil {
			feaset.Open() // 这里面自动判断和处理了
			feaset.Close()
		}
		store.Close()
	}
}

func UpdateSqlite(datafile, indextype string) {
	if len(indextype) == 0 {
		indextype = "xzorder" // 默认xz-order，效果较好
	}
	store := data.NewDatastore(data.StoreSqlite)
	params := data.NewConnParams()
	params["filename"] = datafile
	params["index"] = indextype
	store.Open(params)
	names := store.GetFeasetNames()
	for _, name := range names {
		feaset, _ := store.GetFeasetByName(name)
		if feaset != nil {
			feaset.Open() // 这里面自动判断和处理了
			feaset.Close()
		}
	}
	store.Close()
}

func InputMosaic(txtfile, gmrfile string) {
	var mosaic data.MosaicRaset
	mosaic.Open(txtfile)
	mosaic.Save(gmrfile)
	mosaic.Close()
}

func BuildMosaic(path, gmrfile string) {
	var mosaic data.MosaicRaset
	mosaic.Build(path, gmrfile)
	mosaic.Close()
}
