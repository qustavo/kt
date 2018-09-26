package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gchaincl/kt/k8s"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Table interface {
	Rows() [][]string
}

func RenderTable(table *tview.Table, i Table) *tview.Table {
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
		Select(1, 1).
		SetFixed(1, 0).
		SetSelectionChangedFunc(func(row int, column int) {
			if row == 0 {
				table.Select(row+1, column)
				return
			}
		})

	return table
}

type TableAdapter [][]string

func (t TableAdapter) Rows() [][]string { return t }

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

func main() {
	client, err := k8s.New()
	if err != nil {
		log.Fatal(err)
	}

	deployments, err := client.Deployments()
	if err != nil {
		log.Fatal(err)
	}

	table := tview.NewTable()

	var pods = [][]string{
		[]string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"},
		[]string{"proxy-2423423-23423", "1/1", "Running", "7", "9h"},
		[]string{"a-very-long-name-for-a-fake-service", "1/3", "[:red]Error[white:]", "99", "1d"},
	}

	app := tview.NewApplication()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			RenderTable(table, deployments)
		case tcell.KeyF2:
			RenderTable(table, TableAdapter(pods))
		}

		switch event.Rune() {
		case 'q':
			app.Stop()
		}

		return event
	})

	grid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(0).
		SetBorders(false).
		AddItem(table, 0, 0, 1, 1, 0, 0, true).
		AddItem(menuBar(), 1, 0, 1, 1, 0, 0, false)

	RenderTable(table, deployments)
	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
