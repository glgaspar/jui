package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/glgaspar/jui/data"
	"github.com/rivo/tview"
)

type Project struct {
	Name        string
	BuildNumber string
	App         *tview.Application
	Pages       *tview.Pages
}

type Job struct {
	Name string `json:"name"`
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
	Color string `json:"color"`
	NextBuildNumber int `json:"nextBuildNumber"`
	Builds []struct {
		Number int `json:"number"`
		Result string `json:"result"`
		Timestamp int64 `json:"timestamp"`
	} `json:"builds"`
}

func (pd *Project) ProjectPage() {
	if pd.Name == "" {
		return
	}
	if pd.App == nil || pd.Pages == nil {
		j := &Job{}
		j.FetchJobData(pd.Name)
		fmt.Printf("Job: %s\nDescription: %s\nBuilds: %d\n", j.Name, j.Description, len(j.Builds))
		return
	}

	j := &Job{}
	j.FetchJobData(pd.Name)

	table := tview.NewTable().
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false)

	table.SetCell(0, 0, tview.NewTableCell("Number").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))
	table.SetCell(0, 1, tview.NewTableCell("Result").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))
	table.SetCell(0, 2, tview.NewTableCell("Timestamp").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))

	for i, b := range j.Builds {
		table.SetCell(i+1, 0, tview.NewTableCell(fmt.Sprintf("%d", b.Number)))
		table.SetCell(i+1, 1, tview.NewTableCell(b.Result))
		table.SetCell(i+1, 2, tview.NewTableCell(time.Unix(b.Timestamp/1000, 0).Format("2006-01-02 15:04:05")))
	}

	details := tview.NewTextView().SetDynamicColors(true)
	details.SetBorder(true).SetTitle("Job Details")
	details.SetText(fmt.Sprintf("Name: %s\nDisplayName: %s\nDescription: %v\nColor: %s\nNextBuildNumber: %d",
		j.Name, j.DisplayName, j.Description, j.Color, j.NextBuildNumber))

	logView := tview.NewTextView().SetDynamicColors(true).SetChangedFunc(func() { pd.App.Draw() })
	logView.SetBorder(true).SetTitle("Build Log")

	logView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			logView.ScrollToEnd()
			return nil
		}
		return event
	})

	table.SetSelectedFunc(func(row, column int) {
		if row <= 0 || row > len(j.Builds) {
			return
		}
		b := j.Builds[row-1]
		logText, err := FetchBuildLog(pd.Name, fmt.Sprintf("%d", b.Number))
		if err != nil {
			logView.SetText(fmt.Sprintf("Error fetching log: %v", err))
			return
		}
		logView.SetText(logText)
		pd.Pages.SwitchToPage("project:" + pd.Name)
		pd.App.SetFocus(logView)
	})
	table.SetBorder(true).SetTitle("Builds")

	right := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(details, 7, 0, false).
		AddItem(logView, 0, 1, false)

	layout := tview.NewFlex().
		AddItem(table, 0, 1, true).
		AddItem(right, 0, 2, false)

	layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pd.Pages.RemovePage("project:" + pd.Name)
			pd.App.SetFocus(nil)
			return nil
		}
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'B', 'b':
				pd.App.SetFocus(table)
				return nil
			case 'L', 'l':
				pd.App.SetFocus(logView)
				return nil
			case 'H', 'h':
				pd.Pages.RemovePage("project:" + pd.Name)
				pd.App.SetFocus(nil)
				return nil
			}
		}
		return event
	})

	pd.Pages.AddPage("project:"+pd.Name, layout, true, true)
	pd.App.SetFocus(table)
}

func (pd *Project) BuildLog() {
	if pd.Name == "" || pd.BuildNumber == "" {
		return
	}
	logText, err := FetchBuildLog(pd.Name, pd.BuildNumber)
	if err != nil {
		fmt.Printf("Error fetching log: %v\n", err)
		return
	}
	if pd.App == nil || pd.Pages == nil {
		fmt.Println(logText)
		return
	}

	logView := tview.NewTextView().SetDynamicColors(true).SetText(logText)
	logView.SetBorder(true).SetTitle(fmt.Sprintf("%s #%s", pd.Name, pd.BuildNumber))
	logView.SetDoneFunc(func(key tcell.Key) {
		pd.Pages.RemovePage("buildlog:" + pd.Name + ":" + pd.BuildNumber)
	})
	logView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			logView.ScrollToEnd()
			return nil
		}
		return event
	})
	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(tview.NewBox(), 0, 1, false)
	inner := tview.NewFlex()
	inner.AddItem(tview.NewBox(), 0, 1, false)
	inner.AddItem(logView, 80, 0, true)
	inner.AddItem(tview.NewBox(), 0, 1, false)
	modal.AddItem(inner, 20, 0, true)
	modal.AddItem(tview.NewBox(), 0, 1, false)

	pd.Pages.AddPage("buildlog:"+pd.Name+":"+pd.BuildNumber, modal, true, true)
	pd.App.SetFocus(logView)
}

func (j *Job) FetchJobData(name string) (string, error) {
	res, err := data.Api("GET", fmt.Sprintf("/job/%s/api/json?depth=1", name), nil)
	if err != nil {
		return "", err
	}
	if res == nil {
		return "", fmt.Errorf("empty response")
	}
	if err := json.NewDecoder(bytes.NewReader(*res)).Decode(j); err != nil {
		return "", err
	}
	return string(*res), nil
}

func FetchBuildLog(name string, build string) (string, error) {
	res, err := data.Api("GET", fmt.Sprintf("/job/%s/%s/consoleText", name, build), nil)
	if err != nil {
		return "", err
	}
	if res == nil {
		return "", fmt.Errorf("empty response")
	}

	return string(*res), nil
}
