package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

type Handlers struct {
	collection *mongo.Collection
}

type OHLCVFilter struct {
	Pair      string  `bson:"pair"`
	Timestamp float64 `bson:"timestamp"`
}

func (h Handlers) getData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	params := mux.Vars(r)
	pair := params["pair"]
	// from := params["from"]
	// to := params["to"]

	log.Println(params)

	opts := options.Find()
	opts.SetSort(bson.D{{"timestamp", -1}})
	var dat []bson.M
	// cursor, err := h.collection.Find(ctx, bson.M{"pair": bson.D{{"$eq", pair}}, "timestamp": bson.D{{"$gt", from}, {"$lt", to}}})
	cursor, err := h.collection.Find(ctx, bson.M{"pair": bson.D{{"$eq", pair}}}, opts)
	if err = cursor.All(ctx, &dat); err != nil {
		panic(err)
	}
	if err != nil {
		log.Panicln(err)
	} else {
		// log.Println(dat)
		json.NewEncoder(w).Encode(dat)
	}

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

	return collection, disconnectMongo
}

func main() {
	collection, disconnectMongo := initMongo()
	handler := Handlers{collection: collection}
	defer disconnectMongo()

	r := mux.NewRouter()
	port := os.Getenv("PORT")
	r.HandleFunc("/api/ohlcv", handler.getData).Methods(http.MethodGet).Queries("pair", "{pair}").Queries("from", "{from}").Queries("to", "{to}")
	r.Use(loggingMiddleware)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
