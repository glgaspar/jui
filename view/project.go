package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
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
	Class           string `json:"_class"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	DisplayName     string `json:"displayName"`
	Color           string `json:"color"`
	NextBuildNumber int    `json:"nextBuildNumber"`
	Builds          []struct {
		Number    int    `json:"number"`
		Result    string `json:"result"`
		Timestamp int64  `json:"timestamp"`
	} `json:"builds"`
	Jobs []struct {
		Class string `json:"_class"`
		Name  string `json:"name"`
		URL   string `json:"url"`
		Color string `json:"color"`
	} `json:"jobs"`
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

	if len(j.Jobs) > 0 {
		pd.renderMultiBranchPage(j)
		return
	}

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

	var stopLogPoll chan struct{}
	var stopMutex sync.Mutex

	stopCurrentPoll := func() {
		stopMutex.Lock()
		defer stopMutex.Unlock()
		if stopLogPoll != nil {
			close(stopLogPoll)
			stopLogPoll = nil
		}
	}

	table.SetSelectedFunc(func(row, column int) {
		if row <= 0 || row > len(j.Builds) {
			return
		}
		b := j.Builds[row-1]
		buildNum := fmt.Sprintf("%d", b.Number)

		stopCurrentPoll()

		stopMutex.Lock()
		stopLogPoll = make(chan struct{})
		currentStop := stopLogPoll
		stopMutex.Unlock()

		logText, err := FetchBuildLog(pd.Name, buildNum)
		if err != nil {
			logView.SetText(fmt.Sprintf("Error fetching log: %v", err))
		} else {
			logView.SetText(logText)
		}
		pd.Pages.SwitchToPage("project:" + pd.Name)
		pd.App.SetFocus(logView)

		go func(bNum string, stop chan struct{}) {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					building := IsBuildBuilding(pd.Name, bNum)
					logText, err := FetchBuildLog(pd.Name, bNum)
					pd.App.QueueUpdateDraw(func() {
						if err != nil {
							logView.SetText(fmt.Sprintf("Error fetching log: %v", err))
						} else {
							logView.SetText(logText)
						}
					})
					if !building {
						return
					}
				}
			}
		}(buildNum, currentStop)
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
			stopCurrentPoll()
			pd.Pages.RemovePage("project:" + pd.Name)
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
				stopCurrentPoll()
				pd.Pages.RemovePage("project:" + pd.Name)
				pd.Pages.SwitchToPage("home")
				return nil
			}
		}
		return event
	})

	pd.Pages.AddPage("project:"+pd.Name, layout, true, true)
	pd.App.SetFocus(table)
}

func (pd *Project) renderMultiBranchPage(j *Job) {
	table := tview.NewTable().
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false)

	table.SetCell(0, 0, tview.NewTableCell("Branch Name").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(2))
	table.SetCell(0, 1, tview.NewTableCell("Status").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))

	for i, branch := range j.Jobs {
		table.SetCell(i+1, 0, tview.NewTableCell(branch.Name))
		table.SetCell(i+1, 1, tview.NewTableCell(branch.Color))
	}

	table.SetSelectedFunc(func(row, column int) {
		if row <= 0 || row > len(j.Jobs) {
			return
		}
		branch := j.Jobs[row-1]
		p := &Project{
			Name:  fmt.Sprintf("%s/job/%s", pd.Name, branch.Name),
			App:   pd.App,
			Pages: pd.Pages,
		}
		p.ProjectPage()
	})
	table.SetBorder(true).SetTitle("Branches")

	details := tview.NewTextView().SetDynamicColors(true)
	details.SetBorder(true).SetTitle("Project Details")
	details.SetText(fmt.Sprintf("Name: %s\nDisplayName: %s\nDescription: %v\nBranches: %d",
		j.Name, j.DisplayName, j.Description, len(j.Jobs)))

	right := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(details, 7, 0, false).
		AddItem(tview.NewBox(), 0, 1, false)

	layout := tview.NewFlex().
		AddItem(table, 0, 1, true).
		AddItem(right, 0, 2, false)

	layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pd.Pages.RemovePage("project:" + pd.Name)
			return nil
		}
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'B', 'b':
				pd.App.SetFocus(table)
				return nil
			case 'H', 'h':
				pd.Pages.RemovePage("project:" + pd.Name)
				pd.Pages.SwitchToPage("home")
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

	logView := tview.NewTextView().SetDynamicColors(true).SetText(logText).SetChangedFunc(func() { pd.App.Draw() })
	logView.SetBorder(true).SetTitle(fmt.Sprintf("%s #%s", pd.Name, pd.BuildNumber))

	stopLogPoll := make(chan struct{})
	var stopOnce sync.Once

	logView.SetDoneFunc(func(key tcell.Key) {
		stopOnce.Do(func() { close(stopLogPoll) })
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

	go func(bNum string, stop chan struct{}) {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				building := IsBuildBuilding(pd.Name, bNum)
				logText, err := FetchBuildLog(pd.Name, bNum)
				pd.App.QueueUpdateDraw(func() {
					if err != nil {
						logView.SetText(fmt.Sprintf("Error fetching log: %v", err))
					} else {
						logView.SetText(logText)
					}
				})
				if !building {
					return
				}
			}
		}
	}(pd.BuildNumber, stopLogPoll)
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

func IsBuildBuilding(name string, build string) bool {
	res, err := data.Api("GET", fmt.Sprintf("/job/%s/%s/api/json?tree=building", name, build), nil)
	if err != nil || res == nil {
		return false
	}
	var b struct {
		Building bool `json:"building"`
	}
	if err := json.NewDecoder(bytes.NewReader(*res)).Decode(&b); err != nil {
		return false
	}
	return b.Building
}
