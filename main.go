package main

import (
	"os"

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
	receive()
	// send("599d879410d3150261146e81", "hella from golang")
}
