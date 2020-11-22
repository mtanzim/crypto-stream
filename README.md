# crypto-streams

A simple Go project that receives a stream of cryptocurrency data from a websocket api, and publishes them to a Kafka broker. The messages are then read by a set of clients.

## Components

- [x] A stream producer in Go: subscribes to an external WS stream, and publishes it to a Kafka broker
- [x] A web socket server in node: subscribes to the Kafka stream, and streams it through a websocket connection
- [x] A persister in Go: subscribes to the Kafka stream, and stores the data in MongoDB
- [x] A REST API server in Go: queries the MongoDB and allows access to its data through REST

## To Do

- [ ] Containerize the application with Docker

## Resources

- https://github.com/confluentinc/confluent-kafka-go
- https://kafka.apache.org/quickstart
- https://github.com/gorilla/websocket
- https://exchange.blockchain.com/api/#websocket-api
- https://www.npmjs.com/package/kafkajs
- https://github.com/websockets/ws#simple-server
- https://www.websocket.org/echo.html
