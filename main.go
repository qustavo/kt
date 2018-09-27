package main

import (
	"log"
	"os"

	"github.com/gchaincl/kt/k8s"
)

func main() {
	var ns string
	if len(os.Args) > 1 {
		ns = os.Args[1]
	} else {
		ns = "default"
	}

	client, err := k8s.New(ns)
	if err != nil {
		log.Fatal(err)
	}

	ui := New(client)
	if err := ui.Run(); err != nil {
		log.Fatalln(err)
	}
}
