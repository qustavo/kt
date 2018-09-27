package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Interface interface {
	Deployments() (*Table, error)
	PODs() (*Table, error)
}

type Client struct {
	*kubernetes.Clientset
}

func New() (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	if err != nil {
		return nil, err
	}

	// create the clientset
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &Client{cs}, nil
}

type Table struct {
	rows    [][]string
	updates chan struct{}
}

func NewTable(header []string) *Table {
	ch := make(chan struct{})
	return &Table{
		rows:    [][]string{header},
		updates: ch,
	}
}

func (t *Table) Rows() [][]string       { return t.rows }
func (t *Table) Updates() chan struct{} { return t.updates }
func (t *Table) Push(row []string)      { t.rows = append(t.rows, row) }
func (t *Table) Notify()                { t.updates <- struct{}{} }

func (t *Table) update(fn func([]string) bool, newRow []string) {
	for i, row := range t.rows {
		if fn(row) == true {
			t.rows[i] = newRow
			return
		}
	}
	t.Push(newRow)
}

func (t *Table) delete(fn func(row []string) bool) {
	for i, row := range t.rows {
		if fn(row) == true {
			t.rows = append(t.rows[:i], t.rows[i+1:]...)
			return
		}
	}
}

func (c *Client) Deployments() (*Table, error) {
	t := NewTable(
		[]string{"NAME", "DESIRED", "CURRENT", "UP-TO-DATE", "AVAILABLE", "AGE"},
	)

	opts := meta.ListOptions{}
	watcher, err := c.Apps().Deployments("default").Watch(opts)
	if err != nil {
		return nil, err
	}

	lookUp := func(name string) func([]string) bool {
		return func(row []string) bool {
			if len(row) == 0 {
				return false
			}
			return (name == row[0])
		}
	}

	go func() {
		for {
			ev := <-watcher.ResultChan()
			obj := ev.Object.(*apps.Deployment)
			switch ev.Type {
			case watch.Added:
				t.Push(deploymentRow(obj))
			case watch.Deleted:
				t.delete(lookUp(obj.Name))
			case watch.Modified:
				t.update(lookUp(obj.Name), deploymentRow(obj))
			}
			t.Notify()
		}
	}()

	return t, nil
}

func (c *Client) PODs() (*Table, error) {
	t := NewTable(
		[]string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"},
	)

	opts := meta.ListOptions{}
	watcher, err := c.Core().Pods("default").Watch(opts)
	if err != nil {
		return nil, err
	}

	lookUp := func(name string) func([]string) bool {
		return func(row []string) bool {
			if len(row) == 0 {
				return false
			}
			return (name == row[0])
		}
	}

	go func() {
		for {
			ev := <-watcher.ResultChan()
			obj := ev.Object.(*core.Pod)
			switch ev.Type {
			case watch.Added:
				t.Push(podRow(obj))
			case watch.Deleted:
				t.delete(lookUp(obj.Name))
			case watch.Modified:
				t.update(lookUp(obj.Name), podRow(obj))
			}
			t.Notify()
		}
	}()

	return t, nil
}

func deploymentRow(obj *apps.Deployment) []string {
	return []string{
		obj.Name,
		fmt.Sprintf("%d", *obj.Spec.Replicas),
		fmt.Sprintf("%d", obj.Status.Replicas),
		fmt.Sprintf("%d", obj.Status.UpdatedReplicas),
		fmt.Sprintf("%d", obj.Status.AvailableReplicas),
		formatAge(obj.CreationTimestamp),
	}
}

func podRow(obj *core.Pod) []string {
	var ready int
	for _, c := range obj.Status.ContainerStatuses {
		if c.Ready {
			ready++
		}
	}

	return []string{
		obj.Name,
		fmt.Sprintf("%d/%d", ready, len(obj.Status.ContainerStatuses)),
		string(obj.Status.Phase),
		fmt.Sprintf("%d", 0), //obj.Status.ContainerStatuses[0].RestartCount),
		formatAge(obj.CreationTimestamp),
	}
}

func formatAge(when meta.Time) string {
	diff := time.Now().Sub(when.Time)
	if diff.Hours() > 24 {
		return fmt.Sprintf("%dd", int(diff.Hours())/24)
	}

	if diff.Hours() > 1 {
		return fmt.Sprintf("%dh", int(diff.Hours()))
	}

	if diff.Minutes() > 0 {
		return fmt.Sprintf("%dm", int(diff.Minutes()))
	}

	return fmt.Sprintf("%ds", int(diff.Seconds()))
}
