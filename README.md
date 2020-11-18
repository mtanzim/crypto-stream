# crypto-streams

A simple Go project that receives a stream of cryptocurrency data from a websocket api, and publishes them to a Kafka broker. The messages are then read by a node-js client.

## To Do

- [ ] Build a simple front-end
- [ ] Containerize the application with Docker

## Resources

- https://github.com/confluentinc/confluent-kafka-go
- https://kafka.apache.org/quickstart
- https://github.com/gorilla/websocket
- https://exchange.blockchain.com/api/#websocket-api
- https://www.npmjs.com/package/kafkajs
- https://github.com/websockets/ws#simple-server
- https://www.websocket.org/echo.html
