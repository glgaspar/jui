package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/glgaspar/jui/config"
	"github.com/glgaspar/jui/view"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()

func main() {
	pages := tview.NewPages()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyRune {
			if event.Rune() == '?' {
				view.ShowHelp(app, pages)
				return nil
			}
		}
		return event
	})
	homeView := &view.HomeView{App: app, Pages: pages}

	var homeRendered bool
	configView := &view.ConfigView{
		App:   app,
		Pages: pages,
		OnGoHome: func() {
			if config.APIURL == "" {
				return
			}
			if !homeRendered {
				homeView.Render()
				homeRendered = true
			}
			pages.SwitchToPage("home")
		},
	}
	configView.Render()

	app.SetRoot(pages, true)
	if config.APIURL == "" {
		fmt.Println("Warning: API URL is empty!")
		pages.SwitchToPage("config")
	} else {
		if !homeRendered {
			homeView.Render()
			homeRendered = true
		}
		pages.SwitchToPage("home")
	}
	if err := app.Run(); err != nil {
		fmt.Println("Error running app:", err)
	}
}
