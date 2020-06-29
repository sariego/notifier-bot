package meet

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"sariego.dev/cotalker-bot/base"
	"sariego.dev/cotalker-bot/services/identity"
)

// Driver [Meet] - send meeting links
type Driver struct {
	Client base.Client
}

// meet code pool
var codes = initCodes()
var cursor = 0

// NewMeeting - generates message with meet url
// and notifies @ mentions
func (d Driver) NewMeeting(pkg base.Package) string {
	code := codes[cursor]
	cursor = (cursor + 1) % len(codes)
	msg := fmt.Sprintf(
		"meet.google.com/%v\nhttps://meet.google.com/%v",
		code, code,
	)

	// notify @ mentions
	channels, _ := identity.Driver{Client: d.Client}.
		GetNotifyChannels(pkg)
	sender := identity.GetSenderName(pkg.Author)
	go d.notify(channels, sender, msg)

	return msg
}

func (d Driver) notify(channels []string, sender, msg string) {
	for _, ch := range channels {
		notif := fmt.Sprintf(
			"%v quiere hablar contigo\n%v",
			sender, msg,
		)
		out := base.Package{
			Channel: ch,
			Message: notif,
		}
		d.Client.Send(out)
	}
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
