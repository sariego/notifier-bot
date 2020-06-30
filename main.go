package main

import (
	"os"
	"strings"

	"sariego.dev/cotalker-bot/base"
	"sariego.dev/cotalker-bot/clients/cotalker"
	"sariego.dev/cotalker-bot/services/data"
)

const (
	// VERSION - program version
	VERSION = "0.3.1"
	// NAME - bot name
	NAME = "JAIME"
	// todo docopt
	// USAGE = ``
)

func main() {
	var client base.Client
	client = data.NewCachedClient(&cotalker.Client{}) // todo add other clients
	handler := &pkgHandler{client}

	switch os.Args[1] {
	case "receive":
		client.Receive(handler)
	case "send":
		pkg := base.Package{
			Channel: os.Args[2],
			Message: strings.Join(os.Args[3:], " "),
		}
		client.Send(pkg)
	}
}
