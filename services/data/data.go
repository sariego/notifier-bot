package data

import (
	"database/sql"
	"log"
	"os"

	// load postgres driver
	_ "github.com/lib/pq"
)

// DB - sql connection pool
var DB *sql.DB

func init() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	DB = db
}
