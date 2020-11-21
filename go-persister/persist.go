package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type OHLCV struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Pair      string             `bson:"pair"`
	Timestamp float64            `bson:"timestamp"`
	Open      float64            `bson:"open"`
	High      float64            `bson:"high"`
	Low       float64            `bson:"low"`
	Close     float64            `bson:"close"`
	Volume    float64            `bson:"volume"`
}

type OHLCVFilter struct {
	Pair      string  `bson:"pair"`
	Timestamp float64 `bson:"timestamp"`
}

func initMongo() (*mongo.Collection, func()) {

	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")
	collName := os.Getenv("MONGO_COLL")

	// connect to MongoDB
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Panicln(err)
	}
	disconnectMongo := func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Panicln(err)
		}
	}
	db := client.Database(dbName)
	collection := db.Collection(collName)
	idxModel := mongo.IndexModel{
		Keys: bson.M{
			"pair":      1,
			"timestamp": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	collection.Indexes().CreateOne(ctx, idxModel)

	return collection, disconnectMongo
}

func persistInMongo(collection *mongo.Collection, dat *OHLCV) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// res, err := collection.InsertOne(ctx, dat)
	rplcOpt := options.Replace()
	rplcOpt.SetUpsert(true)

	filter := OHLCVFilter{Timestamp: dat.Timestamp, Pair: dat.Pair}
	log.Println(filter)

	res, err := collection.ReplaceOne(ctx, filter, dat, rplcOpt)
	if err != nil {
		log.Println(err)
	}
	log.Println("Updated ", res.ModifiedCount)
	log.Println("Upserted ", res.UpsertedID)
}

func initKafka() *kafka.Consumer {
	// connect to Kafka
	kafkaServer := os.Getenv("KAFKA_SERVER_ADDR")
	groupId := os.Getenv("KAFKA_GROUP_ID")

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServer,
		"group.id":          groupId,
		"auto.offset.reset": "latest",
	})

	if err != nil {
		panic(err)
	}
	return c
}

func main() {

	c := initKafka()
	defer c.Close()
	pairs := os.Getenv("PAIRS")
	pairsSli := strings.Split(pairs, ",")
	if len(pairsSli) == 0 {
		log.Panicln("No pairs to subscribe to.")
	}
	c.SubscribeTopics(pairsSli, nil)

	collection, disconnectMongo := initMongo()
	defer disconnectMongo()

	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			log.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			var dat OHLCV
			if err := json.Unmarshal(msg.Value, &dat); err != nil {
				log.Println(err)
			}
			go persistInMongo(collection, &dat)
		} else {
			// The client will automatically try to recover from all errors.
			log.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

}
