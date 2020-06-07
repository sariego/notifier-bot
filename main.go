package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	engineio "github.com/googollee/go-engine.io"
	"github.com/googollee/go-engine.io/transport"
	"github.com/googollee/go-engine.io/transport/polling"
	"github.com/googollee/go-engine.io/transport/websocket"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("starting client...")

	host := os.Getenv("COTALKER_HOST")
	token := os.Getenv("COTALKER_BOT_TOKEN")

	url, _ := url.Parse(host + "/socket.io-client/")
	header := http.Header{
		"Authorization": []string{"Bearer " + token},
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
		ft, r, err := conn.NextReader()
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
		fmt.Println("read:", ft, string(b[1:]))
	}

	/* go func() {
		defer conn.Close()
		// listening...
	}() */

	/* for {
		fmt.Println("write text hello")
		w, err := conn.NextWriter(engineio.TEXT)
		if err != nil {
			log.Println("next writer error:", err)
			return
		}
		if _, err := w.Write([]byte("hello")); err != nil {
			w.Close()
			log.Println("write error:", err)
			return
		}
		if err := w.Close(); err != nil {
			log.Println("write close error:", err)
			return
		}
		fmt.Println("write binary 1234")
		w, err = conn.NextWriter(engineio.BINARY)
		if err != nil {
			log.Println("next writer error:", err)
			return
		}
		if _, err := w.Write([]byte{1, 2, 3, 4}); err != nil {
			w.Close()
			log.Println("write error:", err)
			return
		}
		if err := w.Close(); err != nil {
			log.Println("write close error:", err)
			return
		}
		time.Sleep(time.Second * 5)
	} */
}
