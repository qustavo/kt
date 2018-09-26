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

type Deployments struct {
	rows [][]string
}

func (d *Deployments) Rows() [][]string {
	return d.rows
}

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
		diff := time.Now().Sub(item.CreationTimestamp.Time)
		var age string
		if diff.Hours() > 24 {
			age = fmt.Sprintf("%dd", int(diff.Hours())/24)
		} else {
			age = diff.String()
		}

		row := []string{
			item.Name,
			fmt.Sprintf("%d", *item.Spec.Replicas),
			fmt.Sprintf("%d", item.Status.ReadyReplicas),
			fmt.Sprintf("%d", item.Status.UpdatedReplicas),
			fmt.Sprintf("%d", item.Status.AvailableReplicas),
			age,
		}
		d.rows = append(d.rows, row)
	}

	return d, nil
}
