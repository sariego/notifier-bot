package main

import (
	"log"
	"strings"

	"sariego.dev/cotalker-bot/base"
	"sariego.dev/cotalker-bot/services/feedback"
	"sariego.dev/cotalker-bot/services/identity"
	"sariego.dev/cotalker-bot/services/meet"
)

type pkgHandler struct {
	client base.Client
}

func (h *pkgHandler) Handle(pkg base.Package) error {
	split := strings.Split(pkg.Message, " ")

	// handle @<BOTNAME> <CMD> <ARGS>...
	if split[0] == "@"+NAME {
		log.Println("match: detected @ format, switching")
		split = split[1:]
		if split[0][0] != '!' {
			split[0] = "!" + split[0]
		}
	}

	// handle !<CMD> <ARGS>...
	if split[0][0] == '!' {
		log.Println("match: detected ! format")
		cmd := split[0][1:]
		msg, err := execute(instruction{h.client, pkg, cmd, split[1:]})
		if err != nil {
			return err
		}

		if len(msg) > 0 {
			log.Printf("exec: %v@%v\n", cmd, pkg.Channel)
			output := base.Package{
				Channel: pkg.Channel,
				Message: msg,
			}
			h.client.Send(output)
		}
	} else if len(h.client.MentionsRedirectURL()) > 0 {
		// notify mentions
		go identity.Driver{Client: h.client}.NotifyMentions(pkg)
	}

	return nil
}

type instruction struct {
	client base.Client
	pkg    base.Package
	cmd    string
	args   []string
}

// todo markdown responses
func execute(parsed instruction) (response string, err error) {
	if len(parsed.args) == 0 {
		parsed.args = append(parsed.args, "")
	}

	switch parsed.cmd {
	case "ping":
		response = "pong!"
	case "register", "add":
		response, err = identity.Driver{Client: parsed.client}.
			Register(
				parsed.args[0],
				parsed.pkg.Author,
				parsed.pkg.Channel,
			)
	case "deregister", "remove", "delete":
		response, err = identity.Deregister(
			parsed.args[0],
			parsed.pkg.Author,
		)
	case "whoami":
		response, err = identity.WhoAmI(parsed.pkg.Author)
	case "whoishere":
		response, err = identity.Driver{Client: parsed.client}.
			WhoIsHere(parsed.pkg.Channel)
	case "meet":
		response = meet.Driver{Client: parsed.client}.
			NewMeeting(parsed.pkg)
	case "feedback", "bug":
		response, err = feedback.Create(
			parsed.pkg.Author,
			parsed.pkg.Channel,
			parsed.cmd,
			strings.Join(parsed.args, " "),
		)
	}
	return
}
