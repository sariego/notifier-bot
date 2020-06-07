package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	engineio "github.com/googollee/go-engine.io"
	"github.com/googollee/go-engine.io/transport"
	"github.com/googollee/go-engine.io/transport/polling"
	"github.com/googollee/go-engine.io/transport/websocket"
)

type cotalkerEnvelope struct {
	Model   string            `json:"model"`
	Type    string            `json:"type"`
	Count   int               `json:"count"`
	Content []cotalkerMessage `json:"content"`
	Channel []string          `json:"channel"`
}

type cotalkerMessage struct {
	ID          string `json:"_id"`
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	Status      int    `json:"isSaved"`
	Channel     string `json:"channel"`
	Author      string `json:"sentBy"`
}

type cotalkerMultiCMD struct {
	Method  string          `json:"method"`
	Message cotalkerMessage `json:"message"`
}

func receive(handler func(msg string, ch string)) {
	fmt.Println("starting client...")

	url, _ := url.Parse(HOST + "/socket.io-client/")
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

	fmt.Println(conn.ID(), conn.LocalAddr(), "~>", conn.RemoteAddr(), "with", conn.RemoteHeader())

	fmt.Println("listening...")
	for {
		_, r, err := conn.NextReader()
		if err != nil {
			log.Println("error@next_reader:", err)
			return
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			r.Close()
			log.Println("error@read_all:", err)
			return
		}
		if err := r.Close(); err != nil {
			log.Println("error@read_close:", err)
		}
		fmt.Println("bytes:", len(b))
		if len(b) <= 1 {
			continue
		}

		args := strings.SplitN(string(b[2:len(b)-1]), ",", 3) // todo: use reported b[0] count?
		var e cotalkerEnvelope
		err = json.Unmarshal([]byte(args[2]), &e)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("event:", args[0])
		fmt.Println("type:", args[1])
		fmt.Println("subject:", strings.Split(args[1], "#")[0])
		if strings.Split(args[1][1:], "#")[0] != "message" { // hacky hacky
			continue
		}
		msg := e.Content[0].Content
		ch := e.Channel[0]
		log.Printf("read: \"%v\"@%v\n", msg, ch)

		handler(msg, ch)
	}
}

// debug target channel 599d879410d3150261146e81
func send(ch string, msg string) {
	cmd := cotalkerMultiCMD{
		Method: "POST",
		Message: cotalkerMessage{
			ID:          generateCotalkerUUID(),
			Content:     msg,
			ContentType: "text/plain",
			Status:      2,
			Channel:     ch,
			Author:      USERID,
		},
	}
	body := struct {
		CMD []cotalkerMultiCMD `json:"cmd"`
	}{
		CMD: []cotalkerMultiCMD{cmd},
	}
	json, err := json.Marshal(body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("url:", HOST+"/api/messages/multi")
	fmt.Println("json:", string(json))
	req, err := http.NewRequest(http.MethodPost, HOST+"/api/messages/multi", bytes.NewBuffer(json))
	req.Header.Add("Authorization", "Bearer "+TOKEN)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("res: %+v\n", resp)
}

func generateCotalkerUUID() string {
	now := time.Now().Unix()
	rand.Seed(now)
	p0 := fmt.Sprintf("%x", now)
	p1 := USERID[4:8] + USERID[18:20]
	p2 := USERID[20:24]
	p3 := "112233"

	return p0 + p1 + p2 + p3
}
