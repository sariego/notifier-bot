package cotalker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	engineio "github.com/googollee/go-engine.io"
	"github.com/googollee/go-engine.io/transport"
	"github.com/googollee/go-engine.io/transport/polling"
	"github.com/googollee/go-engine.io/transport/websocket"
	"sariego.dev/cotalker-bot/base"
)

var (
	// HOST - cotalker server url
	HOST string = os.Getenv("COTALKER_HOST")
	// WEB - webclient url
	WEB string = os.Getenv("COTALKER_WEB_CLIENT")
	// USERID - cotalker bot user id
	USERID string = os.Getenv("COTALKER_BOT_ID")
	// TOKEN - cotalker bot token
	TOKEN string = os.Getenv("COTALKER_BOT_TOKEN")
)

// Client - cotalker v1 implementation
// uses socket.io to listen on channels
// and api v1 to send messages and fetch data
type Client struct{}

// BotID - returns userid of the bot
func (c *Client) BotID() string {
	return USERID
}

// MentionsRedirectURL - returns link to send users when @ mentioned
// leave empty to disable mentions tracking
func (c *Client) MentionsRedirectURL() string {
	return WEB + "/c/g/summary?channel=%v"
}

// Receive - listens to socket and handles package via handler func
func (c *Client) Receive(handler base.PackageHandler) error {
	log.Println("starting client...")

	url, err := url.Parse(HOST + "/socket.io-client/")
	if err != nil {
		log.Fatalln("error@parse_url:", err)
	}
	header := http.Header{
		"Authorization": []string{"Bearer " + TOKEN},
	}

	dialer := engineio.Dialer{
		Transports: []transport.Transport{polling.Default, websocket.Default},
	}
	conn, err := dialer.Dial(url.String(), header)
	if err != nil {
		log.Fatalln("error@dial:", err)
	}
	defer conn.Close()
	log.Println("conn: ", conn.ID(), conn.LocalAddr(), "~>", conn.RemoteAddr())
	//todo log headers ??
	log.Println("listening...")
	for {
		_, r, err := conn.NextReader()
		if err != nil {
			log.Println("error@next_reader:", err)
			return err
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			r.Close()
			log.Println("error@read_all:", err)
			return err
		}
		if err := r.Close(); err != nil {
			log.Println("error@read_close:", err)
		}
		log.Println("bytes:", len(b))
		if len(b) <= 1 {
			continue
		}

		args := strings.SplitN(string(b[2:len(b)-1]), ",", 3) // todo: use reported b[0] count?
		var e envelope
		err = json.Unmarshal([]byte(args[2]), &e)
		if err != nil {
			log.Println("error@cmd_unmarshal:", err)
		}

		ev := args[0][1 : len(args[0])-1]
		ty := args[1][1 : len(args[1])-1]
		log.Printf("parsed: event:%v type:%v\n", ev, ty)
		if ty != "message#dataArray" {
			continue
		}

		u := e.Content[0].Author
		ch := e.Channel[0]
		msg := e.Content[0].Content
		log.Printf("read: \"%v\"@%v\n", msg, ch)

		pkg := base.Package{
			Author:  u,
			Channel: ch,
			Message: msg,
		}
		go handler.Handle(pkg)
	}
}

// Send - sends message via apiv1 /multi endpoing
func (c *Client) Send(pkg base.Package) error {
	cmd := command{
		Method: "POST",
		Message: message{
			ID:          generateCotalkerUUID(),
			Content:     pkg.Message,
			ContentType: "text/plain",
			Status:      2,
			Channel:     pkg.Channel,
			Author:      USERID,
		},
	}
	body := struct {
		CMD []command `json:"cmd"`
	}{
		CMD: []command{cmd},
	}
	json, err := json.Marshal(body)
	if err != nil {
		log.Println("error@cmd_marshal:", err)
		return err
	}
	log.Printf("send: \"%v\"@%v\n", pkg.Message, pkg.Channel)
	req, err := http.NewRequest(http.MethodPost, HOST+"/api/messages/multi", bytes.NewBuffer(json))
	if err != nil {
		log.Println("error@new_request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+TOKEN)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("error@http_send:", err)
		return err
	}
	defer res.Body.Close()

	// fmt.Printf("req: %+v\n", req)
	// fmt.Printf("res: %+v\n", res)

	return nil
}

// GetChannelInfo - get info from cotalker api
func (c *Client) GetChannelInfo(id string) (base.ChannelInfo, error) {
	req, err := http.NewRequest(http.MethodGet, HOST+"/api/channels/"+id, nil)
	if err != nil {
		log.Println("error@new_request:", err)
		return base.ChannelInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+TOKEN)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("error@http_get:", err)
		return base.ChannelInfo{}, err
	}
	defer res.Body.Close()

	var ch channel
	err = json.NewDecoder(res.Body).Decode(&ch)
	if err != nil {
		log.Println("error@decode_res:", err)
		return base.ChannelInfo{}, err
	}

	return base.ChannelInfo(ch), nil
}

type message struct {
	ID          string `json:"_id"`
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	Status      int    `json:"isSaved"`
	Channel     string `json:"channel"`
	Author      string `json:"sentBy"`
}

type envelope struct {
	Model   string    `json:"model"`
	Type    string    `json:"type"`
	Count   int       `json:"count"`
	Content []message `json:"content"`
	Channel []string  `json:"channel"`
}

type command struct {
	Method  string  `json:"method"`
	Message message `json:"message"`
}

type channel struct {
	ID           string   `json:"_id"`
	Name         string   `json:"nameDisplay"`
	Participants []string `json:"userIds"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// generates compliant cotalker uuid
func generateCotalkerUUID() string {
	p0 := fmt.Sprintf("%08x", time.Now().Unix())
	p1 := USERID[4:8] + USERID[18:20]
	p2 := USERID[20:24]
	p3 := fmt.Sprintf("%06x", rand.Intn(16777216)) // 16^6

	return p0 + p1 + p2 + p3
}
