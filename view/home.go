package view

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type HomeBuildQueueItem struct {
	Class   string `json:"_class"`
	Actions []struct {
		Class  string `json:"_class"`
		Causes []struct {
			Class            string `json:"_class"`
			ShortDescription string `json:"shortDescription"`
		} `json:"causes"`
	} `json:"actions"`
	Blocked      bool   `json:"blocked"`
	Buildable    bool   `json:"buildable"`
	ID           int    `json:"id"`
	InQueueSince int64  `json:"inQueueSince"`
	Params       string `json:"params"`
	Stuck        bool   `json:"stuck"`
	Task         struct {
		Class string `json:"_class"`
		Name  string `json:"name"`
		URL   string `json:"url"`
		Color string `json:"color"`
	} `json:"task"`
	URL                        string `json:"url"`
	Why                        string `json:"why"`
	BuildableStartMilliseconds int64  `json:"buildableStartMilliseconds"`
}

type HomeBuildQueue struct {
	Class             string               `json:"_class"`
	DiscoverableItems []interface{}        `json:"discoverableItems"`
	Items             []HomeBuildQueueItem `json:"items"`
}

func (bq *HomeBuildQueue) FetchBuildQueue() {
	res, err := http.Get(os.Getenv("JENKINS_URL") + "/queue/api/json")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(bq)
	if err != nil {
		panic(err)
	}
}

func (bq *HomeBuildQueue) Display() *tview.Table {
	table := tview.NewTable().
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false)

	table.SetCell(0, 0, tview.NewTableCell("Task Name").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))
	table.SetCell(0, 1, tview.NewTableCell("Reason").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))

	for i, item := range bq.Items {
		table.SetCell(i+1, 0, tview.NewTableCell(item.Task.Name))
		table.SetCell(i+1, 1, tview.NewTableCell(item.Why))
	}
	table.Box.SetBorder(true).SetTitle("Build Queue (1)")

	return table
}

type HomeBuildExecutor struct {
	Class    string `json:"_class"`
	Computer []struct {
		Class       string `json:"_class"`
		DisplayName string `json:"displayName"`
		Executors   []struct {
			CurrentExecutable struct {
				Class   string `json:"_class"`
				Actions []struct {
					Class string `json:"_class,omitempty"`
				} `json:"actions"`
				Artifacts         []interface{} `json:"artifacts"`
				Building          bool          `json:"building"`
				Description       interface{}   `json:"description"`
				DisplayName       string        `json:"displayName"`
				Duration          int           `json:"duration"`
				EstimatedDuration int           `json:"estimatedDuration"`
				Executor          struct {
				} `json:"executor"`
				Fingerprint     []interface{} `json:"fingerprint"`
				FullDisplayName string        `json:"fullDisplayName"`
				ID              string        `json:"id"`
				InProgress      bool          `json:"inProgress"`
				KeepLog         bool          `json:"keepLog"`
				Number          int           `json:"number"`
				QueueID         int           `json:"queueId"`
				Result          interface{}   `json:"result"`
				Timestamp       int64         `json:"timestamp"`
				URL             string        `json:"url"`
				BuiltOn         string        `json:"builtOn"`
				ChangeSet       struct {
					Class string `json:"_class"`
				} `json:"changeSet"`
				Culprits []struct {
				} `json:"culprits"`
			} `json:"currentExecutable"`
		} `json:"executors"`
	} `json:"computer"`
}

func (be *HomeBuildExecutor) FetchBuildExecutors() {
	res, err := http.Get(os.Getenv("JENKINS_URL") + "/computer/api/json?tree=computer[displayName,executors[currentExecutable[*]]]")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(be)
	if err != nil {
		panic(err)
	}
}

func (be *HomeBuildExecutor) Display() *tview.Table {
	table := tview.NewTable().
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false)

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

	return table
}

//color: Indicates build health (e.g., blue=success, red=failure, aborted, disabled, anime=running).
//lastBuild: Details on the most recent build (number, url, result).

type HomeProjectList struct {
	Class string `json:"_class"`
	Jobs  []struct {
		Class     string `json:"_class"`
		Name      string `json:"name"`
		Color     string `json:"color,omitempty"`
		LastBuild *struct {
			Number int    `json:"number"`
			Result string `json:"result"`
		} `json:"lastBuild,omitempty"`
	} `json:"jobs"`
}

func (pl *HomeProjectList) FetchProjectList() {
	res, err := http.Get(os.Getenv("JENKINS_URL") + "/api/json?tree=jobs[name,color,lastBuild[number,result]]")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(pl)
	if err != nil {
		panic(err)
	}
}

func (pl *HomeProjectList) Display() *tview.Table {
	projectDetail := ProjectDetail{}

	table := tview.NewTable().
		// SetBorders(true).
		SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, _ := table.GetSelection()
			if row > 0 && row <= len(pl.Jobs) {
				job := pl.Jobs[row-1]

				projectDetail.Name = job.Name
				projectDetail.TakeOver()

			}
			return nil
		}
		return event
	})
	table.SetSelectedFunc(func(row, column int) {
		// table.GetCell(row, column)

	})

	table.SetCell(0, 0, tview.NewTableCell("Project Name").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(2))
	table.SetCell(0, 1, tview.NewTableCell("Status").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))
	table.SetCell(0, 2, tview.NewTableCell("Last Build").SetAlign(tview.AlignLeft).SetSelectable(false).SetExpansion(1))

	for i, job := range pl.Jobs {
		table.SetCell(i+1, 0, tview.NewTableCell(job.Name))
		table.SetCell(i+1, 1, tview.NewTableCell(job.Color))
		if job.LastBuild != nil {
			buildInfo := fmt.Sprintf("Number: %d, Result: %s", job.LastBuild.Number, job.LastBuild.Result)
			table.SetCell(i+1, 2, tview.NewTableCell(buildInfo))
		} else {
			table.SetCell(i+1, 2, tview.NewTableCell("N/A"))
		}
	}

	table.Box.SetBorder(true).SetTitle("Projects (3)")

	return table
}

type HomeView struct {
	App *tview.Application
}

func (h *HomeView) Render() {
	buildQueue := &HomeBuildQueue{}
	buildQueue.FetchBuildQueue()

	projectList := &HomeProjectList{}
	projectList.FetchProjectList()

	buildExecutor := &HomeBuildExecutor{}
	buildExecutor.FetchBuildExecutors()

	queueTable := buildQueue.Display()
	executorTable := buildExecutor.Display()
	projectTable := projectList.Display()

	focusables := []tview.Primitive{queueTable, executorTable, projectTable}
	focusIndex := 0

	grid := tview.NewGrid().
		SetRows(0, 0).
		SetColumns(0, 0).
		SetBorders(false).
		SetBordersColor(tview.Styles.PrimaryTextColor).
		AddItem(queueTable, 0, 0, 1, 1, 0, 0, true).
		AddItem(executorTable, 1, 0, 1, 1, 0, 0, false).
		AddItem(projectTable, 0, 1, 2, 1, 0, 0, false)

	h.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
			}
		}
		return event
	})	

	h.App.SetRoot(grid, true)
	if err := h.App.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
