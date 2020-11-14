// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// TODO: use config files for symbols/topics/urls etc
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
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

// TODO: clean up
func readMsg(c *websocket.Conn, p *kafka.Producer, done chan struct{}) {
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

		// TODO: fix willy nilly type assertions
		switch {
		case dat["channel"] == "heartbeat":
			log.Println("<3")
		case dat["channel"] == "ticker" && dat["event"] == "snapshot":
			symbol_str := dat["symbol"].(string)
			price_str := strconv.FormatFloat(dat["price_24h"].(float64), 'f', -1, 64)
			volume_str := strconv.FormatFloat(dat["volume_24h"].(float64), 'f', -1, 64)
			str_slice := []string{symbol_str, price_str, volume_str}
			log.Println(strings.Join(str_slice, ","))
		case dat["channel"] == "prices" && dat["event"] == "updated":
			// pass message to kafka
			// topic := "quickstart-events"
			topic := dat["symbol"].(string)
			// log.Println(topic)
			p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Value:          []byte(message),
			}, nil)
		default:
			log.Printf("recv: %s", message)

		}
	}
}

// kafka delivery report handler
func kafkaDeliveryReports(p *kafka.Producer) {
	for e := range p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
			} else {
				fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
			}
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

	// connect kafka
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		log.Panicln(err)
	}
	go kafkaDeliveryReports(p)
	defer p.Close()

	done := make(chan struct{})
	subcribeHeartbeat(c)
	symbols := []string{"BTC-USD", "BTC-EUR", "ETH-USD", "ETH-EUR"}
	for _, s := range symbols {
		subscribeOHLCV(c, s)
	}

	go readMsg(c, p, done)
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
