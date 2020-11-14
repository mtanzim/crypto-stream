// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "ws.prod.blockchain.info", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/mercury-gateway/v1/ws"}
	log.Printf("connecting to %s", u.String())

	header := http.Header{}
	header.Set("origin", "https://exchange.blockchain.com")

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	c.WriteMessage(websocket.TextMessage, []byte(`
	{
		"action": "subscribe",
		"channel": "heartbeat"
	  }
	`))
	c.WriteMessage(websocket.TextMessage, []byte(`
	{
		"action": "subscribe",
		"channel": "ticker",
		"symbol": "BTC-USD"
	  }
	`))
	c.WriteMessage(websocket.TextMessage, []byte(`
	{
		"action": "subscribe",
		"channel": "ticker",
		"symbol": "ETH-USD"
	  }
	`))

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

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
