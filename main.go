package main

import (
	"flag"
	"log"
	"os"

	"github.com/gchaincl/kt/k8s"
)

func main() {
	// We use NewFlagSet because plog polutes the global flags.
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ns := fs.String("n", "default", "Kubernetes Namespace")
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	client, err := k8s.New(*ns)
	if err != nil {
		log.Fatal(err)
	}

	ui := New(client)
	if err := ui.Run(); err != nil {
		log.Fatalln(err)
	}
}
