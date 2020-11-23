# crypto-streams

A Go/Node project that receives a stream of cryptocurrency data from an external websocket api, and publishes them to a Kafka broker. The messages are then subscribed to by a set of clients.

## Components

- [x] A stream producer in Go: subscribes to an external WS stream, reformats it, and publishes it to a Kafka broker
- [x] A web socket server in Node: subscribes to the Kafka stream, and serves the stream through a websocket connection
- [x] A persister in Go: subscribes to the Kafka stream, and persists the data in MongoDB
- [x] A REST API server in Go: queries the MongoDB and allows access to its data through a REST API

## To Do

- [x] Containerize the application with Docker
- [ ] Move `.env` config to `docker-compose.yml`

## Note

- Configure the pairs through `.env` files

## Example Websocket Server Request

- Connect to `ws://localhost:3004`
- Send the following message: `subscribe,BTC-USD`

### Example Response

`RECEIVED: {"pair":"BTC-USD","timestamp":1606166220000,"open":18383.38,"high":18385.85,"low":18376.035,"close":18382.06,"volume":0.002185}`

## Example REST API Request

### Request

`GET http://localhost:8080/api/ohlcv?pair=BTC-EUR&from=0&to=9999999999999`

### Response

```json
[
  {
    "_id": "5fbc23c061728e44120173cc",
    "close": 15551.51,
    "high": 15551.91,
    "low": 15541.255,
    "open": 15541.255,
    "pair": "BTC-EUR",
    "timestamp": 1606165440000,
    "volume": 0
  },
  {
    "_id": "5fbc238461728e4412016c67",
    "close": 15541.255,
    "high": 15544.99,
    "low": 15534.435,
    "open": 15535.985,
    "pair": "BTC-EUR",
    "timestamp": 1606165380000,
    "volume": 0
  },
  {
    "_id": "5fbc234861728e4412016665",
    "close": 15535.985,
    "high": 15546.085,
    "low": 15535.985,
    "open": 15538.6,
    "pair": "BTC-EUR",
    "timestamp": 1606165320000,
    "volume": 0
  },
  {
    "_id": "5fbc233d61728e441201627d",
    "close": 15538.6,
    "high": 15548.665,
    "low": 15537.84,
    "open": 15540.59,
    "pair": "BTC-EUR",
    "timestamp": 1606165260000,
    "volume": 0
  },
  {
    "_id": "5fbc22d05dd4eb1fefa26db4",
    "close": 15547.44,
    "high": 15551.9,
    "low": 15547.44,
    "open": 15548.94,
    "pair": "BTC-EUR",
    "timestamp": 1606165200000,
    "volume": 0
  },
  {
    "_id": "5fbc22945dd4eb1fefa26562",
    "close": 15548.94,
    "high": 15559.89,
    "low": 15547.54,
    "open": 15551.845,
    "pair": "BTC-EUR",
    "timestamp": 1606165140000,
    "volume": 0.04699165
  }
]
```

## Resources

- https://exchange.blockchain.com/api/#websocket-api
- https://github.com/confluentinc/confluent-kafka-go
- https://kafka.apache.org/quickstart
- https://www.npmjs.com/package/kafkajs
- https://github.com/gorilla/websocket
- https://github.com/websockets/ws#simple-server
- https://github.com/bitnami/bitnami-docker-kafka/blob/master/README.md
- https://www.websocket.org/echo.html
