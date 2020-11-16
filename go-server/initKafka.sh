# Start the ZooKeeper service
bin/zookeeper-server-start.sh config/zookeeper.properties &
bin/kafka-server-start.sh config/server.properties &

bin/kafka-topics.sh --create --topic BTC-USD--bootstrap-server localhost:9092
bin/kafka-topics.sh --create --topic BTC-EUR --bootstrap-server localhost:9092
bin/kafka-topics.sh --create --topic ETH-USD --bootstrap-server localhost:9092
bin/kafka-topics.sh --create --topic ETH-EUR --bootstrap-server localhost:9092