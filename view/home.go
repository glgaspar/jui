package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/glgaspar/jui/data"
	"github.com/rivo/tview"
)

type HomeBuildQueueItem struct {
	Task struct {
		Name string `json:"name"`
	} `json:"task"`
	Why string `json:"why"`
}

type HomeBuildQueue struct {
	Items []HomeBuildQueueItem `json:"items"`
}

func (bq *HomeBuildQueue) FetchBuildQueue() error {
	res, err := data.Api("GET", "/queue/api/json", nil)
	if err != nil {
		return err
	}
	err = json.NewDecoder(bytes.NewReader(*res)).Decode(bq)
	if err != nil {
		return err
	}
	return nil
}

func (bq *HomeBuildQueue) Display() *tview.Table {
	table := tview.NewTable().
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false)

	bq.UpdateTable(table)

	return table
}

func (bq *HomeBuildQueue) UpdateTable(table *tview.Table) {
	table.Clear()
	table.SetCell(0, 0, tview.NewTableCell("Task Name").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))
	table.SetCell(0, 1, tview.NewTableCell("Reason").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))

	for i, item := range bq.Items {
		table.SetCell(i+1, 0, tview.NewTableCell(item.Task.Name))
		table.SetCell(i+1, 1, tview.NewTableCell(item.Why))
	}
	table.Box.SetBorder(true).SetTitle("Build Queue (1)")
}

type HomeBuildExecutor struct {
	Computer []struct {
		DisplayName string `json:"displayName"`
		Executors   []struct {
			CurrentExecutable struct {
				FullDisplayName string      `json:"fullDisplayName"`
				Result          interface{} `json:"result"`
			} `json:"currentExecutable"`
		} `json:"executors"`
	} `json:"computer"`
}

func (be *HomeBuildExecutor) FetchBuildExecutors() error {
	res, err := data.Api("GET", "/computer/api/json?tree=computer[displayName,executors[currentExecutable[*]]]", nil)
	if err != nil {
		return err
	}
	err = json.NewDecoder(bytes.NewReader(*res)).Decode(be)
	if err != nil {
		return err
	}
	return nil
}

func (be *HomeBuildExecutor) Display() *tview.Table {
	table := tview.NewTable().
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false)

	be.UpdateTable(table)

	return table
}

func (be *HomeBuildExecutor) UpdateTable(table *tview.Table) {
	table.Clear()
	table.SetCell(0, 0, tview.NewTableCell("Executor").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))
	table.SetCell(0, 1, tview.NewTableCell("Task").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))
	table.SetCell(0, 2, tview.NewTableCell("Status").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))

	rowIndexer := 1
	for _, computer := range be.Computer {
		for _, executor := range computer.Executors {
			if executor.CurrentExecutable.FullDisplayName == "" {
				continue
			}
			table.SetCell(rowIndexer, 0, tview.NewTableCell(computer.DisplayName))
			table.SetCell(rowIndexer, 1, tview.NewTableCell(executor.CurrentExecutable.FullDisplayName))
			resultStr := "Building"
			if res, ok := executor.CurrentExecutable.Result.(string); ok {
				resultStr = res
			}
			table.SetCell(rowIndexer, 2, tview.NewTableCell(resultStr))
			rowIndexer++
		}
	}
	table.Box.SetBorder(true).SetTitle("Build Executors (2)")
}

//color: Indicates build health (e.g., blue=success, red=failure, aborted, disabled, anime=running).
//lastBuild: Details on the most recent build (number, url, result).

type HomeProjectList struct {
	Jobs []struct {
		Name      string `json:"name"`
		Color     string `json:"color,omitempty"`
		LastBuild *struct {
			Number int    `json:"number"`
			Result string `json:"result"`
		} `json:"lastBuild,omitempty"`
	} `json:"jobs"`
	Filter       bool
	FilterString string
}

func (pl *HomeProjectList) FetchProjectList() {
	res, err := data.Api("GET", "/api/json?tree=jobs[name,color,lastBuild[number,result]]", nil)
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(bytes.NewReader(*res)).Decode(pl)
	if err != nil {
		panic(err)
	}
}

func (pl *HomeProjectList) UpdateTable(table *tview.Table) {
	table.Clear()
	table.SetCell(0, 0, tview.NewTableCell("Project Name").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(2))
	table.SetCell(0, 1, tview.NewTableCell("Status").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))
	table.SetCell(0, 2, tview.NewTableCell("Last Build").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))

	var filteredJobs []struct {
		Name      string `json:"name"`
		Color     string `json:"color,omitempty"`
		LastBuild *struct {
			Number int    `json:"number"`
			Result string `json:"result"`
		} `json:"lastBuild,omitempty"`
	}

	if len(pl.FilterString) > 0 {
		filterLower := strings.ToLower(pl.FilterString)
		for _, job := range pl.Jobs {
			if strings.Contains(strings.ToLower(job.Name), filterLower) {
				filteredJobs = append(filteredJobs, job)
			}
		}
	} else {
		filteredJobs = pl.Jobs
	}

	for i, job := range filteredJobs {
		table.SetCell(i+1, 0, tview.NewTableCell(job.Name).SetReference(job.Name))
		table.SetCell(i+1, 1, tview.NewTableCell(job.Color))
		if job.LastBuild != nil {
			buildInfo := fmt.Sprintf("Number: %d, Result: %s", job.LastBuild.Number, job.LastBuild.Result)
			table.SetCell(i+1, 2, tview.NewTableCell(buildInfo))
		} else {
			table.SetCell(i+1, 2, tview.NewTableCell("N/A"))
		}
	}
	table.Select(1, 0)
}

func (pl *HomeProjectList) Display() (*tview.Flex, *tview.Table, *tview.InputField) {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	filter := tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault)

	table := tview.NewTable().
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false)

	pl.UpdateTable(table)

	filter.SetChangedFunc(func(text string) {
		pl.FilterString = text
		pl.UpdateTable(table)
	})

	flex.AddItem(table, 0, 1, true)

	if pl.Filter {
		flex.AddItem(filter, 1, 0, false)
	}

	return flex, table, filter
}

type HomeView struct {
	App   *tview.Application
	Pages *tview.Pages
}

func (h *HomeView) Render() {
	buildQueue := &HomeBuildQueue{}
	_ = buildQueue.FetchBuildQueue()

	projectList := &HomeProjectList{}
	projectList.FetchProjectList()

	buildExecutor := &HomeBuildExecutor{}
	_ = buildExecutor.FetchBuildExecutors()

	queueTable := buildQueue.Display()
	executorTable := buildExecutor.Display()
	projectFlex, projectTable, projectFilter := projectList.Display()

	projectFlex.SetBorder(true).SetTitle("Projects (3)")

	queueTable.SetFocusFunc(func() { queueTable.SetBorderColor(tcell.ColorGreen) })
	queueTable.SetBlurFunc(func() { queueTable.SetBorderColor(tcell.ColorWhite) })

	executorTable.SetFocusFunc(func() { executorTable.SetBorderColor(tcell.ColorGreen) })
	executorTable.SetBlurFunc(func() { executorTable.SetBorderColor(tcell.ColorWhite) })

	projectTable.SetFocusFunc(func() { projectFlex.SetBorderColor(tcell.ColorGreen) })
	projectTable.SetBlurFunc(func() { projectFlex.SetBorderColor(tcell.ColorWhite) })
	projectFilter.SetFocusFunc(func() { projectFlex.SetBorderColor(tcell.ColorGreen) })
	projectFilter.SetBlurFunc(func() { projectFlex.SetBorderColor(tcell.ColorWhite) })

	queueTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, _ := queueTable.GetSelection()
			if row > 0 && row <= len(buildQueue.Items) {
				item := buildQueue.Items[row-1]
				p := &Project{Name: item.Task.Name, App: h.App, Pages: h.Pages}
				p.ProjectPage()
			}
			return nil
		}
		return event
	})

	executorTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, _ := executorTable.GetSelection()

			var name string
			currentRow := 1
			for _, computer := range buildExecutor.Computer {
				for _, executor := range computer.Executors {
					if executor.CurrentExecutable.FullDisplayName == "" {
						continue
					}
					if currentRow == row {
						full := executor.CurrentExecutable.FullDisplayName
						parts := strings.SplitN(full, " #", 2)
						name = parts[0]
					}
					currentRow++
				}
			}
			if name != "" {
				p := &Project{Name: name, App: h.App, Pages: h.Pages}
				p.ProjectPage()
			}
			return nil
		}
		return event
	})

	projectTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, _ := projectTable.GetSelection()
			if row > 0 {
				cell := projectTable.GetCell(row, 0)
				if cell != nil && cell.GetReference() != nil {
					name := cell.GetReference().(string)
					p := &Project{Name: name, App: h.App, Pages: h.Pages}
					p.ProjectPage()
				}
			}
			return nil
		}
		if event.Rune() == '/' {
			projectList.Filter = true
			projectFlex.AddItem(projectFilter, 1, 0, false)
			h.App.SetFocus(projectFilter)
			return nil
		}
		return event
	})

	projectFilter.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter, tcell.KeyEscape:
			projectList.Filter = false
			projectFlex.RemoveItem(projectFilter)
			h.App.SetFocus(projectTable)
		}
	})

	focusables := []tview.Primitive{queueTable, executorTable, projectTable}
	focusIndex := 0

	grid := tview.NewGrid().
		SetRows(0, 0).
		SetColumns(0, 0).
		SetBorders(false).
		SetBordersColor(tview.Styles.PrimaryTextColor).
		AddItem(queueTable, 0, 0, 1, 1, 0, 0, true).
		AddItem(executorTable, 1, 0, 1, 1, 0, 0, false).
		AddItem(projectFlex, 0, 1, 2, 1, 0, 0, false)

	grid.SetFocusFunc(func() {
		h.App.SetFocus(focusables[focusIndex])
	})

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			focusIndex = (focusIndex + 1) % len(focusables)
			h.App.SetFocus(focusables[focusIndex])
			return nil
		}
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case '1':
				h.App.SetFocus(queueTable)
				focusIndex = 0
				return nil
			case '2':
				h.App.SetFocus(executorTable)
				focusIndex = 1
				return nil
			case '3':
				h.App.SetFocus(projectTable)
				focusIndex = 2
				return nil
			case 'q':
				h.App.Stop()
				return nil
			case 'C':
				h.Pages.SwitchToPage("config")
				return nil
			}
		}
		return event
	})

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			newBuildQueue := &HomeBuildQueue{}
			errQ := newBuildQueue.FetchBuildQueue()

			newBuildExecutor := &HomeBuildExecutor{}
			errE := newBuildExecutor.FetchBuildExecutors()

			h.App.QueueUpdateDraw(func() {
				if errQ == nil {
					buildQueue.Items = newBuildQueue.Items
					buildQueue.UpdateTable(queueTable)
				}
				if errE == nil {
					buildExecutor.Computer = newBuildExecutor.Computer
					buildExecutor.UpdateTable(executorTable)
				}
			})
		}
	}()

	h.Pages.AddPage("home", grid, true, true)
}
