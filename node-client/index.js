require("dotenv").config();

const { Kafka } = require("kafkajs");
const http = require("http");
const WebSocket = require("ws");

const kafka = new Kafka({
  clientId: "my-app",
  brokers: ["localhost:9092"],
});

const port = process.env.PORT || 3004;
const server = http.createServer();
server.listen(port);

const wss = new WebSocket.Server({ noServer: true });
wss.on("connection", function connection(ws, request, client) {
  ws.on("message", function message(msg) {
    console.log(`Received message ${msg} from user ${client}`);
    if (msg === "subscribe") {
      ws.send("Subscribed");
    }
  });
});

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


const consumer = kafka.consumer({ groupId: process.env.KAFKA_GROUP_ID });

const run = async (topics) => {
  await consumer.connect();
  await Promise.all(
    topics.map((topic) => {
      return consumer.subscribe({ topic, fromBeginning: false });
    })
  );

  await consumer.run({
    eachMessage: async ({ _topic, _partition, message }) => {
      const value = JSON.parse(message.value.toString());
      console.log({
        value,
      });
      wss.clients.forEach((client) => {
        if (client.readyState === WebSocket.OPEN) {
          client.send(message.value.toString());
        }
      });
    },
  });
};

const pairs = (process.env.PAIRS || "").split(",");
run(pairs).catch(console.error);
