package main

import (
	"log"
	"strings"

	"sariego.dev/cotalker-bot/base"
	"sariego.dev/cotalker-bot/clients/cotalker"
	"sariego.dev/cotalker-bot/services/meet"
)

func main() {
	client := cotalker.Client{}

	client.Receive(func(pkg base.Package) {
		cmd := strings.Split(pkg.Message, " ")
		if cmd[0][0] == '!' {
			out := base.Package{Channel: pkg.Channel}
			switch cmd[0][1:] {
			case "ping":
				log.Printf("exec: PING@%v\n", pkg.Channel)
				out.Message = "pong!"
			case "meet":
				log.Printf("exec: MEET@%v\n", pkg.Channel)
				out.Message = meet.Respond()
			}
			if out.Message != "" {
				client.Send(out)
			}
		}
	})
}
