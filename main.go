package main

import (
	"os"
	"strings"

	"sariego.dev/notifier-bot/base"
	"sariego.dev/notifier-bot/clients/cotalker"
	"sariego.dev/notifier-bot/services/data"
)

const (
	// VERSION - program version
	VERSION = "0.5.0"
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
