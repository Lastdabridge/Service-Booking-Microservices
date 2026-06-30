package broker

import (
	"log"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func NewTopic(topicName string) (*kafka.Conn, *kafka.Conn) {
	conn, err := kafka.Dial("tcp", "kafka:9092")
	if err != nil {
		log.Fatal(err)
	}

	controller, err := conn.Controller()
	if err != nil {
		log.Fatal(err)
	}

	controllerConn, err := kafka.Dial(
		"tcp",
		controller.Host+":"+strconv.Itoa(controller.Port),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = controllerConn.CreateTopics(
		kafka.TopicConfig{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	return conn, controllerConn
}
