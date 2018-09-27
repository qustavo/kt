package main

import (
	"testing"
	"time"

	"github.com/gchaincl/kt/k8s"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

type client struct {
	sync chan struct{}
}

func (c *client) Deployments() (*k8s.Table, error) {
	t := k8s.NewTable(
		[]string{"NAME", "DESIRED", "CURRENT", "UP-TO-DATE", "AVAILABLE", "AGE"},
	)

	t.Push([]string{"a1", "b1", "c1", "d1", "e1", "f1"})
	go func() {
		t.Notify()
		c.sync <- struct{}{}
	}()

	return t, nil
}

func (c *client) PODs() (*k8s.Table, error) {
	t := k8s.NewTable(
		[]string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"},
	)

	t.Push([]string{"a3", "b3", "c3", "d3", "e3"})
	go func() {
		t.Notify()
		c.sync <- struct{}{}
	}()

	return t, nil
}

func TestUI(t *testing.T) {
	sync := make(chan struct{})
	c := &client{sync}

	sim := tcell.NewSimulationScreen("")
	ui := New(c)
	ui.app.SetScreen(sim)
	go ui.Run()

	// wait for deployments and pods to complete updating
	<-sync
	<-sync

	var bytes []byte
	cells, _, _ := sim.GetContents()
	for _, cell := range cells {
		bytes = append(bytes, cell.Bytes...)
	}
	assert.Contains(t, string(bytes), "UP-TO-DATE")

	sim.InjectKey(tcell.KeyF2, ' ', tcell.ModNone)
	time.Sleep(100 * time.Millisecond)

	bytes = []byte{}
	cells, _, _ = sim.GetContents()
	for _, cell := range cells {
		bytes = append(bytes, cell.Bytes...)
	}
	assert.Contains(t, string(bytes), "READY")
}
