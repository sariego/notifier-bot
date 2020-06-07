package main

import (
	"log"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

var (
	// HOST cotalker server url
	HOST string = os.Getenv("COTALKER_HOST")
	// USERID cotalker bot user id
	USERID string = os.Getenv("COTALKER_BOT_ID")
	// TOKEN cotalker bot token
	TOKEN string = os.Getenv("COTALKER_BOT_TOKEN")
)

func main() {
	receive(func(msg, ch string) {
		cmd := strings.Split(msg, " ")
		if cmd[0][0] == '!' {
			switch cmd[0][1:] {
			case "ping":
				log.Printf("exec: PING@%v\n", ch)
				send(ch, "pong!")
			case "meet":
				log.Printf("exec: MEET@%v\n", ch)
				send(ch, meet())
			}
		}
	})
}
