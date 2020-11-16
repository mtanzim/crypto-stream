// +build ignore

package main

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

func main() {

	// TODO: move to config files
	kafkaServer := os.Getenv("KAFKA_SERVER_ADDR")
	groupId := os.Getenv("KAFKA_GROUP_ID")

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServer,
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})

	defer c.Close()
	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{"BTC-USD", "ETH-USD"}, nil)

	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

}
