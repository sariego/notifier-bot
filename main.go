package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"sariego.dev/cotalker-bot/base"
	"sariego.dev/cotalker-bot/clients/cotalker"
	"sariego.dev/cotalker-bot/services/meet"
)

// todo docopt

func main() {
	var client base.Client
	client = cotalker.Client{}
	// todo add other clients

	switch os.Args[1] {
	case "receive":
		client.Receive(func(pkg base.Package) {
			split := strings.Split(pkg.Message, " ")
			if split[0][0] == '!' {
				cmd := split[0][1:]
				output := base.Package{
					Channel: pkg.Channel,
					Message: generateResponse(cmd, split[1:]),
				}
				if output.Message != "" {
					log.Printf("exec: %v@%v\n", cmd, pkg.Channel)
					client.Send(output)
				}
			}
		})
	case "send":
		pkg := base.Package{
			Channel: os.Args[2],
			Message: strings.Join(os.Args[3:], " "),
		}
		client.Send(pkg)
	}

}

func generateResponse(cmd string, args []string) (response string) {
	switch cmd {
	case "ping":
		response = "pong!"
	case "meet":
		response = meet.Respond()
	case "test":
		response = testDB()
	}
	return
}

func testDB() string {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	var username string
	err = db.QueryRow("select username from identity where user_id = 'user1'").Scan(&username)
	if err != nil {
		log.Fatal(err)
	}
	return username
}
