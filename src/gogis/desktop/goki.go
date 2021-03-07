package desktop

import (
	"fmt"
	"gogis/base"
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

func ShowKI(gmap *mapping.Map, width, height int) {
	// gimain.
	gimain.Main(func() {
		mainrunbasic(gmap, width, height)
	})
}

func mainrunbasic(gmap *mapping.Map, width, height int) {

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

	win.SetCloseCleanFunc(func(w *gi.Window) {
		go gi.Quit() // once main window is closed, quit
	})

	win.MainMenuUpdated()
	vp.UpdateEndNoSig(updt)

	go drawMap(win.Viewport.Pixels, gmap)

	// Ticker保管一个通道，并每隔一段时间向其传递"tick"。
	ticker := time.NewTicker(500 * time.Millisecond)
	// 用Ticker实现定时器
	for {
		select {
		case <-ticker.C:
			vp.ReRender2DTree()
		}
	}
}

func drawMap(img *image.RGBA, gmap *mapping.Map) {
	tr := base.NewTimeRecorder()
	gmap.PrepareImage(img)
	gmap.Draw()
	tr.Output("draw map")
}
