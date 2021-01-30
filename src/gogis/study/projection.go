package main

// "github.com/go-spatial/proj"
import (
	"fmt"
	"gogis/base"

	"github.com/go-spatial/proj"
)

func prjmain() {
	// testLoad()
	testproj()
	// testParse()
	fmt.Println("DONE!")
}

func testLoad() {
	// var goProj data.Proj
	// goProj.Load()
}

func testParse() {
	// EPSG3395                    EPSGCode = 3395
	// WorldMercator                        = EPSG3395
	// EPSG3857                             = 3857
	// WebMercator                          = EPSG3857
	// EPSG4087                             = 4087
	// WorldEquidistantCylindrical          = EPSG4087
	// EPSG4326                             = 4326
	// WGS84                                = EPSG4326

	// var projStrings = map[EPSGCode]string{
	// 	EPSG 3395: "+proj=merc +lon_0=0 +k=1 +x_0=0 +y_0=0 +datum=WGS84",                            // TODO: support +units=m +no_defs
	// 	EPSG 3857: "+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0", // TODO: support +units=m +nadgrids=@null +wktext +no_defs
	// 	EPSG 4087: "+proj=eqc +lat_ts=0 +lat_0=0 +lon_0=0 +x_0=0 +y_0=0 +datum=WGS84",               // TODO: support +units=m +no_defs
	// }
	// GEOGCS["GCS_WGS_1984",DATUM["D_WGS_1984",SPHEROID["WGS_1984",6378137,298.257223563]],PRIMEM["Greenwich",0],UNIT["Degree",0.017453292519943295]]

	epsg := 2964
	// projString := "+proj=aea +lat_1=55 +lat_2=65 +lat_0=50 +lon_0=-154 +x_0=0 +y_0=0 +datum=NAD27 +units=us-ft +no_defs"
	// projString := "+proj=merc +lon_0=0 +k=1 +x_0=0 +y_0=0 +datum=WGS84 +units=m +no_defs"
	projString := "+proj=merc +lon_0=0 +k=1 +x_0=0 +y_0=0 +datum=WGS84 +units=us-ft +no_defs"
	proj.Register(proj.EPSGCode(epsg), projString)

	pnts := []float64{-98, 39, 98, 39}
	// pnts := []float64{11, 22}
	res, err := proj.Convert(proj.EPSGCode(epsg), pnts)
	if err != nil {
		fmt.Println("proj error:", err)
	}
	fmt.Println("res:", res)
	res2, _ := proj.Inverse(proj.EPSGCode(epsg), res)
	fmt.Println("res2:", res2)
}

func testproj() {
	projInfo := base.PrjFromEpsg(4326)
	proj.Register(proj.EPSGCode(projInfo.Epsg), projInfo.Proj4)
	pnts := []float64{11, 22}
	res, err := proj.Convert(proj.EPSG4326, pnts)
	if err != nil {
		fmt.Println("proj error:", err)
	}
	fmt.Println("res:", res)
	res2, _ := proj.Inverse(proj.EPSG4326, res)
	fmt.Println("res2:", res2)
}
