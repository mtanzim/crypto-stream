require("dotenv").config();

const { Kafka } = require("kafkajs");
const http = require("http");
const WebSocket = require("ws");
const crypto = require("crypto");

const brokers = (process.env.KAFKA_BROKERS || "").split(",");

const kafka = new Kafka({
  clientId: "my-app",
  brokers,
});

const subscriptions = new Map();
const pairs = (process.env.PAIRS || "").split(",");
const pairsSet = new Set(pairs);
const port = process.env.PORT || 3004;

// HTTP server
const server = http.createServer();
server.listen(port);
// TODO: fix auth!!!
const authenticate = (_request, cb) => {
  cb(null, "test");
};
server.on("upgrade", function upgrade(request, socket, head) {
  authenticate(request, (err, client) => {
    if (err || !client) {
      socket.write("HTTP/1.1 401 Unauthorized\r\n\r\n");
      socket.destroy();
      return;
    }

    wss.handleUpgrade(request, socket, head, function done(ws) {
      wss.emit("connection", ws, request, client);
    });
  });
});

// Websocket server
const wss = new WebSocket.Server({ noServer: true });
wss.on("connection", function connection(ws, _request, _client) {
  const clientId = crypto.randomBytes(16).toString("hex");
  subscriptions.set(clientId, { pairs: new Set(), conn: ws });
  ws.on("message", function message(msg) {
    console.log(`Received message ${msg} from user ${clientId}`);
    const [cmd, pair] = msg.split(",");
    if (cmd === "subscribe" && pairsSet.has(pair)) {
      const cur = subscriptions.get(clientId);
      const update = { ...cur, pairs: new Set(cur.pairs.add(pair)) };
      subscriptions.set(clientId, update);
      ws.send(`${clientId} subscribed to ${pair}`);
    }
  });
});

// Kafka client
const consumer = kafka.consumer({ groupId: process.env.KAFKA_GROUP_ID });
const run = async (topics) => {
  await consumer.connect();
  await Promise.all(
    topics.map((topic) => {
      return consumer.subscribe({ topic, fromBeginning: false });
    })
  );

  await consumer.run({
    eachMessage: async ({ topic, _partition, message }) => {
      const value = JSON.parse(message.value.toString());
      console.log({
        value,
      });
      subscriptions.forEach((sub) => {
        const { conn, pairs } = sub;
        if (pairs.has(topic)) {
          conn.send(message.value.toString());
        }
      });
    },
  });
};
// Main
run(pairs).catch(console.error);
