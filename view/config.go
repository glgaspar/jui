package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/glgaspar/jui/config"
	"github.com/rivo/tview"
)

func init() {

}

type ConfigView struct {
	App      *tview.Application
	Pages    *tview.Pages
	OnGoHome func()
}

func (cv *ConfigView) Render() {
	urlView := cv.URLDisplay(cv.Pages)
	table := cv.HeadersDisplay(cv.Pages)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(urlView, 3, 1, true).
		AddItem(table, 0, 2, false)

	// Intercept Tab to toggle focus between URL text view and Headers table
	layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if urlView.HasFocus() {
				cv.App.SetFocus(table)
			} else {
				cv.App.SetFocus(urlView)
			}
			return nil
		}
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'H':
				if cv.OnGoHome != nil {
					cv.OnGoHome()
				} else {
					cv.Pages.SwitchToPage("home")
				}
				return nil
			}
		}
		return event
	})

	// Visual indicators for which panel is currently focused
	urlView.SetFocusFunc(func() {
		urlView.SetBorderColor(tcell.ColorGreen)
	})
	urlView.SetBlurFunc(func() {
		urlView.SetBorderColor(tcell.ColorWhite)
	})
	table.SetFocusFunc(func() {
		table.SetBorderColor(tcell.ColorGreen)
	})
	table.SetBlurFunc(func() {
		table.SetBorderColor(tcell.ColorWhite)
	})

	// Center the configuration view in a small overlay box
	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(tview.NewBox(), 0, 1, false)
	inner := tview.NewFlex()
	inner.AddItem(tview.NewBox(), 0, 1, false)
	inner.AddItem(layout, 80, 0, true)
	inner.AddItem(tview.NewBox(), 0, 1, false)
	modal.AddItem(inner, 20, 0, true)
	modal.AddItem(tview.NewBox(), 0, 1, false)

	cv.Pages.AddPage("config", modal, true, true)
}

func (cv *ConfigView) URLDisplay(pages *tview.Pages) *tview.TextView {
	tv := tview.NewTextView().SetText(config.APIURL)
	tv.SetBorder(true).SetTitle("API URL (Enter to edit)")

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			cv.URLShowForm(pages, tv)
			return nil
		}
		return event
	})

	return tv
}

func (cv *ConfigView) URLShowForm(pages *tview.Pages, tv *tview.TextView) {
	form := tview.NewForm()
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cv.Pages.RemovePage("url_modal")
			cv.App.SetFocus(tv)
			return nil
		}
		return event
	})

	urlField := tview.NewInputField().SetLabel("API URL: ").SetText(config.APIURL).SetFieldWidth(50)
	form.AddFormItem(urlField)
	form.AddButton("Save", func() {
		newUrl := urlField.GetText()
		config.SaveApi(newUrl)
		tv.SetText(newUrl)
		cv.Pages.RemovePage("url_modal")
		cv.App.SetFocus(tv)
	})
	form.AddButton("Cancel", func() {
		cv.Pages.RemovePage("url_modal")
		cv.App.SetFocus(tv)
	})

	form.SetBorder(true).SetTitle("Edit API URL")

	// Center the form in a box overlay
	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(tview.NewBox(), 0, 1, false)
	inner := tview.NewFlex()
	inner.AddItem(tview.NewBox(), 0, 1, false)
	inner.AddItem(form, 60, 0, true)
	inner.AddItem(tview.NewBox(), 0, 1, false)
	modal.AddItem(inner, 10, 0, true)
	modal.AddItem(tview.NewBox(), 0, 1, false)

	cv.Pages.AddPage("url_modal", modal, true, true)
	cv.App.SetFocus(form)
}

func HeaderShowForm(title, key, value string, cv *ConfigView, table *tview.Table, onSave func(newKey, newVal string)) {
	form := tview.NewForm()
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cv.Pages.RemovePage("modal")
			cv.App.SetFocus(table)
			return nil
		}
		return event
	})

	keyField := tview.NewInputField().SetLabel("Key: ").SetText(key)
	valField := tview.NewInputField().SetLabel("Value: ").SetText(value)
	form.AddFormItem(keyField)
	form.AddFormItem(valField)
	form.AddButton("Save", func() {
		newKey := keyField.GetText()
		newVal := valField.GetText()
		if newKey == "" {
			return
		}
		onSave(newKey, newVal)
		cv.Pages.RemovePage("modal")
		cv.App.SetFocus(table)
	})
	form.AddButton("Cancel", func() {
		cv.Pages.RemovePage("modal")
		cv.App.SetFocus(table)
	})
	form.SetBorder(true).SetTitle(title)

	// center the form in a box
	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(tview.NewBox(), 0, 1, false)
	inner := tview.NewFlex()
	inner.AddItem(tview.NewBox(), 0, 1, false)
	inner.AddItem(form, 60, 0, true)
	inner.AddItem(tview.NewBox(), 0, 1, false)
	modal.AddItem(inner, 10, 0, true)
	modal.AddItem(tview.NewBox(), 0, 1, false)

	cv.Pages.AddPage("modal", modal, true, true)
	cv.App.SetFocus(form)
}

func HeadersMountTable(table *tview.Table) {
	table.Clear()
	table.SetCell(0, 0, tview.NewTableCell("KEY").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))
	table.SetCell(0, 1, tview.NewTableCell("VALUE").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))
	i := 0
	for k, v := range config.APIHEADERS {
		table.SetCell(i+1, 0, tview.NewTableCell(k))
		table.SetCell(i+1, 1, tview.NewTableCell(v))
		i++
	}
	table.Box.SetBorder(true).SetTitle("Headers (e=edit d=delete a=add)")

}

func HeadersPrepare() *tview.Table {
	table := tview.NewTable().
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false)

	return table
}

func (cv *ConfigView) HeadersDisplay(pages *tview.Pages) *tview.Table {
	table := HeadersPrepare()

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			r := event.Rune()
			switch r {
			case 'e', 'E':
				row, _ := table.GetSelection()
				if row <= 0 {
					break
				}
				oldKey := table.GetCell(row, 0).Text
				oldVal := table.GetCell(row, 1).Text
				HeaderShowForm("Edit Header", oldKey, oldVal, cv, table, func(newKey, newVal string) {
					headers := map[string]string{}
					for k, v := range config.APIHEADERS {
						headers[k] = v
					}
					if newKey != oldKey {
						delete(headers, oldKey)
					}
					headers[newKey] = newVal
					if err := config.SaveHeaders(headers); err != nil {
						return
					}
					HeadersMountTable(table)
				})
			case 'a', 'A':
				HeaderShowForm("Add Header", "", "", cv, table, func(newKey, newVal string) {
					headers := map[string]string{}
					for k, v := range config.APIHEADERS {
						headers[k] = v
					}
					headers[newKey] = newVal
					if err := config.SaveHeaders(headers); err != nil {
						return
					}
					HeadersMountTable(table)
				})
			case 'd', 'D':
				row, _ := table.GetSelection()
				if row <= 0 {
					break
				}
				key := table.GetCell(row, 0).Text
				headers := map[string]string{}
				for k, v := range config.APIHEADERS {
					headers[k] = v
				}
				delete(headers, key)
				if err := config.SaveHeaders(headers); err != nil {
					return nil
				}
				HeadersMountTable(table)
			}
		}
		return event
	})

	HeadersMountTable(table)

	return table
}
