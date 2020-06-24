package meet

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

// meet code pool
var codes = initCodes()
var cursor = 0

// Respond - generates complete message with meet url
func Respond() string {
	msg := fmt.Sprintf(
		"meet.google.com/%v\nhttps://meet.google.com/%v",
		codes[cursor],
		codes[cursor],
	)
	cursor = (cursor + 1) % len(codes)
	return msg
}

func initCodes() []string {
	// read from meet.dat
	// todo refactor into postgres
	b, err := ioutil.ReadFile("services/meet/codes.dat")
	if err != nil {
		log.Fatalln("fatal_error@read_codes_from_file:", err)
	}

	// remove final empty line if present
	s := strings.Split(string(b), "\n")
	if strings.TrimSpace(s[len(s)-1]) == "" {
		s = s[:len(s)-1]
	}
	return s
}
