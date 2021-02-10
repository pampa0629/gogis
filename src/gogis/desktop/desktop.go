// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"gogis/base"
	"gogis/data"
	_ "gogis/data/memory"
	_ "gogis/data/shape"
	_ "gogis/data/sqlite"
	"gogis/mapping"
	"image"
	"time"

	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/ki/ki"

	"github.com/goki/gi/gist"
	"github.com/goki/gi/units"
)

func main() {
	gimain.Main(func() {
		mainrunbasic()
	})
}

func mainrunbasic() {
	width := 1024
	height := 768

	rec := ki.Node{}          // receiver for events
	rec.InitName(&rec, "rec") // this is essential for root objects not owned by other Ki tree nodes

	win := gi.NewMainWindow("gogis", "gogis desktop", width, height)
	vp := win.WinViewport2D()
	updt := vp.UpdateStart()
	mfr := win.SetMainFrame()
	mfr.WidgetSig.Connect(rec.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Println("mfr.WidgetSig.Connect")
	})

	rlay := gi.AddNewFrame(mfr, "rowlay", gi.LayoutHoriz)
	rlay.SetProp("text-align", "center")
	rlay.SetStretchMaxWidth()

	brow := gi.AddNewLayout(mfr, "brow", gi.LayoutHoriz)
	brow.SetProp("spacing", units.NewEx(2))
	brow.SetProp("horizontal-align", gist.AlignLeft)
	// brow.SetProp("horizontal-align", gi.AlignJustify)
	brow.SetStretchMaxWidth()

	button1 := gi.AddNewButton(brow, "button1")
	button1.SetProp("#icon", ki.Props{ // note: must come before SetIcon
		"width":  units.NewEm(1.5),
		"height": units.NewEm(1.5),
	})
	button1.Tooltip = "press this <i>button</i> to pop up a dialog box"

	icnm := "wedge-down"
	button1.SetIcon(icnm)
	button1.ButtonSig.Connect(rec.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received button signal: %v from button: %v\n", gi.ButtonSignals(sig), send.Name())
		if sig == int64(gi.ButtonClicked) { // note: 3 diff ButtonSig sig's possible -- important to check
			// vp.Win.Quit()
			fmt.Printf("gi.ButtonClicked")
			gi.StringPromptDialog(vp, "", "Enter value here..",
				gi.DlgOpts{Title: "Button1 Dialog", Prompt: "This is a string prompt dialog!  Various specific types of dialogs are available."},
				rec.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					dlg := send.(*gi.Dialog)
					if sig == int64(gi.DialogAccepted) {
						val := gi.StringPromptDialogValue(dlg)
						fmt.Printf("got string value: %v\n", val)
					}
				})
		}
	})

	// lbl := gi.AddNewLabel(rlay, "label1", "This is test text")
	// lbl.SetProp("text-align", "center")
	// lbl.SetProp("border-width", 1)
	// lbl.SetStretchMaxWidth()

	// edit1 := gi.AddNewTextField(rlay, "edit1")
	// button1 := gi.AddNewButton(rlay, "button1")
	// button2 := gi.AddNewButton(rlay, "button2")
	// slider1 := gi.AddNewSlider(rlay, "slider1")
	// spin1 := gi.AddNewSpinBox(rlay, "spin1")

	// edit1.SetText("Edit this text")
	// edit1.SetProp("min-width", "20em")
	// button1.Text = "Button 1"
	// button2.Text = "Button 2"
	// slider1.Dim = gi.X
	// slider1.SetProp("width", "20em")
	// slider1.SetValue(0.5)
	// spin1.SetValue(0.0)

	// main menu
	appnm := gi.AppName()
	mmen := win.MainMenu
	mmen.ConfigMenus([]string{appnm, "Edit", "Window"})

	amen := win.MainMenu.ChildByName(appnm, 0).(*gi.Action)
	amen.Menu = make(gi.Menu, 0, 10)
	amen.Menu.AddAppMenu(win)

	emen := win.MainMenu.ChildByName("Edit", 1).(*gi.Action)
	emen.Menu = make(gi.Menu, 0, 10)
	emen.Menu.AddCopyCutPaste(win)

	win.SetCloseCleanFunc(func(w *gi.Window) {
		go gi.Quit() // once main window is closed, quit
	})

	win.MainMenuUpdated()
	vp.UpdateEndNoSig(updt)

	// go drawMap(win.Viewport.Pixels)
	go drawTiff(win.Viewport.Pixels)

	// Ticker保管一个通道，并每隔一段时间向其传递"tick"。
	ticker := time.NewTicker(500 * time.Millisecond)
	// 用Ticker实现定时器
	for {
		select {
		case <-ticker.C:
			fmt.Println("Ticker...")
			vp.ReRender2DTree()
		}
	}
	// win.StartEventLoop()

	// var input string
	// fmt.Scanln(&input)
}

var gTitle = "JBNTBHTB" // chinapnt_84 point2 JBNTBHTB

func drawMap(img *image.RGBA) {
	tr := base.NewTimeRecorder()
	store := data.NewDatastore(data.StoreShape)
	params := data.NewConnParams()
	params["filename"] = "C:/temp/" + gTitle + ".shp" // sqlite udbx shp
	params["cache"] = true
	params["fields"] = []string{}
	store.Open(params)
	feaset := store.GetFeasetByNum(0)
	feaset.Open()
	tr.Output("open data")

	gmap := mapping.NewMap()
	// var theme mapping.RangeTheme // UniqueTheme
	// gmap.AddLayer(feaset, &theme)
	gmap.AddLayer(feaset, nil)
	gmap.PrepareImage(img)
	// gmap.Zoom(10)
	gmap.Draw()
	tr.Output("draw map")
}

func drawTiff(img *image.RGBA) {
	tr := base.NewTimeRecorder()

	// title := "raster" // JBNTBHTB chinapnt_84
	filename := "C:/temp/filelist.txt"

	var raset data.MosaicRaset
	raset.Open(filename)

	tr.Output("open")

	gmap := mapping.NewMap()
	gmap.AddRasterLayer(raset)

	// gmap.Prepare(1024, 768)
	gmap.PrepareImage(img)
	gmap.Zoom(2)
	gmap.Draw()

	// 输出图片文件
	tr.Output("draw map")
	// gmap.Output2File(gPath+"image.jpg", "jpg")
	// tr.Output("save picture file")

}
