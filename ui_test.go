package main

import (
	"testing"
	"time"

	"github.com/gchaincl/kt/k8s"
)

type client struct{}

func (c *client) Deployments() (*k8s.Table, error) {
	t := k8s.NewTable(
		[]string{"NAME", "DESIRED", "CURRENT", "UP-TO-DATE", "AVAILABLE", "AGE"},
	)

	t.Push([]string{"a1", "b1", "c1", "d1", "e1", "f1"})

	go func() {
		for {
			t.Notify()
			time.Sleep(1 * time.Second)
		}
	}()

	return t, nil
}

func (c *client) PODs() (*k8s.Table, error) {
	t := k8s.NewTable(
		[]string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"},
	)

	go func() {
		for {
			t.Notify()
			t.Push([]string{"a3", "b3", "c3", "d3", "e3"})
			time.Sleep(1 * time.Second)
		}
	}()

	return t, nil
}

func TestUI(t *testing.T) {
	ui := New(&client{})
	ui.Run()
}
