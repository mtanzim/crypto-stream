require("dotenv").config();
const { Kafka } = require("kafkajs");

const kafka = new Kafka({
  clientId: "my-app",
  brokers: ["localhost:9092"],
});

const consumer = kafka.consumer({ groupId: process.env.KAFKA_GROUP_ID });

const run = async (topics: string[]) => {
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
    },
  });
};

const pairs = (process.env.PAIRS || "").split(",");
run(pairs).catch(console.error);
