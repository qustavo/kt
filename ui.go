package main

import (
	"fmt"
	"strings"

	"github.com/gchaincl/kt/k8s"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Table interface {
	Rows() [][]string
	Updates() chan struct{}
}

func renderTable(table *tview.Table, i Table) {
	table.Clear()

	for i, row := range i.Rows() {
		for j, column := range row {
			cell := tview.NewTableCell(column)
			if i == 0 {
				cell.SetBackgroundColor(tcell.ColorGreen)
				cell.SetTextColor(tcell.ColorBlack)
			}

			table.SetCell(i, j, cell)
		}
	}

	table.
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectionChangedFunc(func(row int, column int) {
			if row == 0 {
				table.Select(row+1, column)
				return
			}
		})

	x, y := table.GetSelection()
	if x == 0 && y == 0 {
		table.Select(1, 1)
	}
}

func menuBar() *tview.TextView {
	actions := []string{
		"F1 Deployments",
		"F2 PODs",
		"Q Quit",
	}

	for i, action := range actions {
		split := strings.SplitN(action, " ", 2)
		actions[i] = fmt.Sprintf("%s[black:aqua]%s[white:-]", split[0], split[1])
	}

	return tview.NewTextView().SetDynamicColors(true).SetText(strings.Join(actions, " "))
}

const (
	DeploymentsTable = iota
	PODsTable
)

type UI struct {
	app    *tview.Application
	table  *tview.Table
	client k8s.Interface

	// State
	activeTable int
	deployments Table
	pods        Table
}

func New(c k8s.Interface) *UI {
	return &UI{
		app:    tview.NewApplication(),
		table:  tview.NewTable(),
		client: c,
	}
}

func (ui *UI) setupApp() {
	grid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(0).
		SetBorders(false).
		AddItem(ui.table, 0, 0, 1, 1, 0, 0, true).
		AddItem(menuBar(), 1, 0, 1, 1, 0, 0, false)

	ui.app.
		SetRoot(grid, true).
		SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			switch ev.Key() {
			case tcell.KeyF1:
				renderTable(ui.table, ui.deployments)
				ui.activeTable = DeploymentsTable
			case tcell.KeyF2:
				renderTable(ui.table, ui.pods)
				ui.activeTable = PODsTable
			}

			switch ev.Rune() {
			case 'q':
				ui.app.Stop()
			}

			return ev
		})
}

func (ui *UI) Run() error {
	var err error
	ui.setupApp()

	ui.deployments, err = ui.client.Deployments()
	if err != nil {
		return err
	}
	renderTable(ui.table, ui.deployments)

	ui.pods, err = ui.client.PODs()
	if err != nil {
		return err
	}

	go ui.watch()
	return ui.app.Run()
}

func (ui *UI) watch() {
	for {
		// wait events on any source to trigger the app redraw.
		select {
		case <-ui.deployments.Updates():
		case <-ui.pods.Updates():
		}

		switch ui.activeTable {
		case DeploymentsTable:
			renderTable(ui.table, ui.deployments)
		case PODsTable:
			renderTable(ui.table, ui.pods)
		}

		ui.app.Draw()
	}
}
