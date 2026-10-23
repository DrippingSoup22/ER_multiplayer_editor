// Command ermultiplayereditor is the editor's desktop window.
package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/DrippingSoup22/ER_multiplayer_editor/internal/ui"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("ER Multiplayer Editor"), app.Size(unit.Dp(1180), unit.Dp(760)), app.MinSize(unit.Dp(900), unit.Dp(560)))
		state := ui.New(w)
		if path := os.Getenv("ER_EDITOR_SAVE"); path != "" {
			state.Load(path)
		}
		var ops op.Ops
		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				if e.Err != nil {
					log.Fatal(e.Err)
				}
				os.Exit(0)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				state.Layout(gtx)
				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}
