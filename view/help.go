package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func ShowHelp(app *tview.Application, pages *tview.Pages) {
	helpText := `Help — Global and Page Controls

Global:
  ?: Open this help popup
  Esc: Close popup / modal
  Tab: Cycle focus between panels
  q: Quit application (home)

Home Page:
  1 / 2 / 3: Focus Queue / Executors / Projects
  Enter: Open selected item (Project page)
  C: Open Config

Config Page:
  Enter: Edit API URL
  Tab: Toggle between URL and Headers
  e: Edit selected header
  a: Add header
  d: Delete selected header
  H: Go to Home

Project Page:
  Esc: Go to home
  H: Go to Home
  B: Focus builds table
  L: Focus build log
  On Builds table:
    Enter: Open selected build log
  On Build Log:
	Enter: Scroll to bottom

Press Esc to close this help.
`

	text := tview.NewTextView().SetText(helpText).SetWordWrap(true)
	text.SetBorder(true).SetTitle("Help (?) — press Esc to close")

	text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.RemovePage("help")
			// app.SetFocus(nil)
			return nil
		}
		return event
	})

	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(tview.NewBox(), 0, 1, false)
	inner := tview.NewFlex()
	inner.AddItem(tview.NewBox(), 0, 1, false)
	inner.AddItem(text, 80, 0, true)
	inner.AddItem(tview.NewBox(), 0, 1, false)
	modal.AddItem(inner, 20, 0, true)
	modal.AddItem(tview.NewBox(), 0, 1, false)

	pages.AddPage("help", modal, true, true)
	app.SetFocus(text)
}
