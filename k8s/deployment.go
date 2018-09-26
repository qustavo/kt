package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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

type Deployments struct{ rows [][]string }

func (d *Deployments) Rows() [][]string { return d.rows }

func (c *Client) Deployments() (*Deployments, error) {
	list, err := c.Apps().Deployments("default").List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	d := &Deployments{
		rows: [][]string{
			[]string{"NAME", "DESIRED", "CURRENT", "UP-TO-DATE", "AVAILABLE", "AGE"},
		},
	}

	for _, item := range list.Items {
		row := []string{
			item.Name,
			fmt.Sprintf("%d", *item.Spec.Replicas),
			fmt.Sprintf("%d", item.Status.ReadyReplicas),
			fmt.Sprintf("%d", item.Status.UpdatedReplicas),
			fmt.Sprintf("%d", item.Status.AvailableReplicas),
			formatAge(item.CreationTimestamp),
		}
		d.rows = append(d.rows, row)
	}

	return d, nil
}

type PODs struct{ rows [][]string }

func (p *PODs) Rows() [][]string { return p.rows }

func (c *Client) PODs() (*PODs, error) {
	list, err := c.Core().Pods("default").List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	p := &PODs{
		rows: [][]string{
			[]string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"},
		},
	}

	for _, item := range list.Items {
		var ready int
		for _, c := range item.Status.ContainerStatuses {
			if c.Ready {
				ready++
			}
		}

		row := []string{
			item.Name,
			fmt.Sprintf("%d/%d", ready, len(item.Status.ContainerStatuses)),
			string(item.Status.Phase),
			fmt.Sprintf("%d", item.Status.ContainerStatuses[0].RestartCount),
			formatAge(item.CreationTimestamp),
		}
		p.rows = append(p.rows, row)
	}

	return p, nil
}

func formatAge(when meta.Time) string {
	diff := time.Now().Sub(when.Time)
	if diff.Hours() > 24 {
		return fmt.Sprintf("%dd", int(diff.Hours())/24)
	}

	if diff.Hours() > 1 {
		return fmt.Sprintf("%dh", int(diff.Hours()))
	}

	return diff.String()
}
