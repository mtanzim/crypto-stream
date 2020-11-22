package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

func connect() *websocket.Conn {
	scheme := os.Getenv("WS_SCHEME")
	wsAddr := os.Getenv("WS_ADDR")
	path := os.Getenv("WS_PATH")
	origin := os.Getenv("WS_ORIGIN_HEADER")

	var addr = flag.String("addr", wsAddr, "Blockchain Exchange")
	u := url.URL{Scheme: scheme, Host: *addr, Path: path}
	log.Printf("connecting to %s", u.String())

	header := http.Header{}
	header.Set("origin", origin)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	return c
}

const HearbeatMsg = `
{
	"action": "subscribe",
	"channel": "heartbeat"
  }
`

func subcribeHeartbeat(c *websocket.Conn) {
	c.WriteMessage(websocket.TextMessage, []byte(HearbeatMsg))
}

func makeOHLCVMsg(pair string) string {
	ohlcvStr :=
		`{
		"action": "subscribe",
		"channel": "prices",
		"granularity": 60,
		"symbol": "` + pair + `"
	}`

	return ohlcvStr
}

func subscribeOHLCV(c *websocket.Conn, pair string) {
	c.WriteMessage(websocket.TextMessage, []byte(makeOHLCVMsg(pair)))
}

const ChannelKey = "channel"
const ChannelValPrices = "prices"
const ChannelValHeartbeat = "heartbeat"
const EventKey = "event"
const EventValOHLCV = "updated"
const OHLCVKey = "price"
const SymbolKey = "symbol"

type OHLCV struct {
	Seqnum  int       `json:"seqnum"`
	Event   string    `json:"event"`
	Channel string    `json:"channel"`
	Symbol  string    `json:"symbol"`
	Price   []float64 `json:"price"`
}

type OHLCVConverted struct {
	Pair      string  `json:"pair"`
	Timestamp float64 `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

func (ohlcv OHLCV) reformat() ([]byte, *OHLCVConverted) {
	converted := OHLCVConverted{
		Pair: ohlcv.Symbol}

	for i, v := range ohlcv.Price {
		switch i {
		case 0:
			converted.Timestamp = v
		case 1:
			converted.Open = v
		case 2:
			converted.High = v
		case 3:
			converted.Low = v
		case 4:
			converted.Close = v
		case 5:
			converted.Volume = v
		}
	}
	convertedJSON, err := json.Marshal(converted)
	if err != nil {
		log.Println(err)
	}
	return convertedJSON, &converted
}

func sendToKafka(p *kafka.Producer, message []byte) {
	ohlcv := OHLCV{}
	if err := json.Unmarshal(message, &ohlcv); err != nil {
		log.Println(err)
	}
	convertedJSON, converted := ohlcv.reformat()
	// pass message to kafka
	topic := converted.Pair
	err := p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          convertedJSON,
	}, nil)
	if err != nil {
		log.Panicln(err)
	}

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

		log.Printf("recv: %s", message)

		switch {
		case dat[ChannelKey] == ChannelValHeartbeat:
			log.Println("<3")
		case dat[ChannelKey] == ChannelValPrices && dat[EventKey] == EventValOHLCV:
			sendToKafka(p, message)

		}
	}
}

// kafka delivery report handler
func kafkaDeliveryReports(p *kafka.Producer) {
	for e := range p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				log.Printf("Delivery failed: %v\n", ev.TopicPartition)
			} else {
				log.Printf("Delivered message to %v\n", ev.TopicPartition)
			}
		default:
			log.Println(ev)
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
	kafkaServer := os.Getenv("KAFKA_SERVER_ADDR")
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaServer})
	if err != nil {
		log.Fatalln(err)
	}
	go kafkaDeliveryReports(p)
	defer p.Close()

	done := make(chan struct{})
	subcribeHeartbeat(c)
	pairsEnv := os.Getenv("PAIRS")
	pairs := strings.Split(pairsEnv, ",")
	for _, pair := range pairs {
		subscribeOHLCV(c, pair)
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
