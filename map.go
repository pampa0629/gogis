package main

import (
	"fmt"

	"gogis/base"
	"gogis/data"
	_ "gogis/data/memory"
	"gogis/draw"
	"gogis/mapping"
)

// 默认的进度条
func defProgress(title, sub string, no, count int, step, total int64, cost, estimate int) bool {
	str := fmt.Sprintf("%s;%s;\tpro:%d/%d;\tstep:%d/%d;\ttime cost:%d,remain:%d",
		title, sub, no, count, step, total, cost, estimate)
	fmt.Println(str)
	return false
}

func DrawMap(mapfile, picfile string, width, height, batch int) {
	mapping.BATCH_COUNT = batch
	gmap := mapping.NewMap(draw.Default)
	gmap.Open(mapfile)
	gmap.Prepare(width, height)
	gmap.Draw()
	gmap.Output2File(picfile, draw.MapType(base.GetExt(picfile)))
	gmap.Close()
}

func Cache(mapfile, maptype, cachepath string) {
	gmap := mapping.NewMap(draw.Default)
	gmap.Open(mapfile)
	mapTile := mapping.NewMapTile(gmap, mapping.Epsg4326)
	mapTile.Cache(cachepath, gmap.Name, draw.MapType(maptype), defProgress)
}

func add2Map(gmap *mapping.Map, datafile, name string) {
	ext := base.GetExt(datafile)
	var store data.Datastore
	switch ext {
	case "shp":
		store = data.NewDatastore(data.StoreShape)
	case "sqlite", "udbx":
		store = data.NewDatastore(data.StoreSqlite)
	}
	if store != nil {
		params := data.NewConnParams()
		params["filename"] = datafile
		store.Open(params)
		feaset, _ := store.GetFeasetByName(name)
		if feaset == nil {
			feaset = store.GetFeasetByNum(0)
		}
		feaset.Open()
		gmap.AddFeatureLayer(feaset, nil)
	} else if ext == "gmr" {
		var raset data.MosaicRaset
		raset.Open(datafile)
		gmap.AddRasterLayer(&raset)
	}
}

func CreateMap(mapfile, datafile, name string) {
	gmap := mapping.NewMap(draw.Default)
	gmap.Open(mapfile)
	add2Map(gmap, datafile, name)
	gmap.Save(mapfile)
	gmap.Close()
}
