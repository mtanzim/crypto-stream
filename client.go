// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

func connect() *websocket.Conn {
	var addr = flag.String("addr", "ws.prod.blockchain.info", "Blockchain Exchange")
	u := url.URL{Scheme: "wss", Host: *addr, Path: "/mercury-gateway/v1/ws"}
	log.Printf("connecting to %s", u.String())

	header := http.Header{}
	header.Set("origin", "https://exchange.blockchain.com")

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	return c
}

func subcribeHeartbeat(c *websocket.Conn) {
	c.WriteMessage(websocket.TextMessage, []byte(`
	{
		"action": "subscribe",
		"channel": "heartbeat"
	  }
	`))
}

func subscribeOHLCV(c *websocket.Conn, symbol string) {
	c.WriteMessage(websocket.TextMessage, []byte(`
	{
		"action": "subscribe",
		"channel": "prices",
		"granularity": 60,
		"symbol": "`+symbol+`"
	  }
	`))
}

func readMsg(c *websocket.Conn, done chan struct{}) {
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		var dat map[string]interface{}
		if err := json.Unmarshal(message, &dat); err != nil {
			log.Println(err)
		}
		switch dat["channel"] {
		case "heartbeat":
			log.Println("<3")
		case "ticker":
			if dat["event"].(string) == "snapshot" {
				symbol_str := dat["symbol"].(string)
				price_str := strconv.FormatFloat(dat["price_24h"].(float64), 'f', -1, 64)
				volume_str := strconv.FormatFloat(dat["volume_24h"].(float64), 'f', -1, 64)
				str_slice := []string{symbol_str, price_str, volume_str}

				log.Println(strings.Join(str_slice, ","))
			}
		default:
			log.Printf("recv: %s", message)

		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c := connect()
	defer c.Close()

	done := make(chan struct{})
	subcribeHeartbeat(c)
	symbols := []string{"BTC-USD", "BTC-EUR", "ETH-USD", "ETH-EUR"}
	for _, s := range symbols {
		subscribeOHLCV(c, s)
	}

	go readMsg(c, done)
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}

			return
		}
	}
}
