package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/glgaspar/jui/view"
	"github.com/joho/godotenv"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		return event
	})

	homeView := &view.HomeView{App: app}
	homeView.Render()

	// all this is just me trying to figure out how to use tview, will be removed later
	// box1 := tview.NewBox().SetBorder(false).SetTitle("JUI")
	// box2 := tview.NewBox().SetBorder(false).SetTitle("JUI")

	// grid := tview.NewGrid().
	// 	SetRows(1,0).
	// 	SetBorders(true).
	// 	AddItem(box1, 0, 0, 1, 1, 0, 0, false).
	// 	AddItem(box2, 1, 0, 1, 1, 0, 0, false)
	// 	// AddItem(tview.NewBox(), 1, 2, 1, 1, 0, 100, false)

	// app.SetTitle("JUI").
	// 	SetRoot(grid, true).
	// 	SetFocus(box1)

	// if err := app.SetRoot(grid, true).Run(); err != nil {
	// 	panic(err)
	// }
}
