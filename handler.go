package main

import (
	"log"
	"strings"

	"sariego.dev/cotalker-bot/base"
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
		response, err = identity.Register(
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
	case "meet":
		response = meet.Respond()
	}
	return
}

