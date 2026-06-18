package broker

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func createTopic() error {
	conn, err := kafka.Dial("tcp", fmt.Sprintf("%s", os.Getenv("KAFKA_BROKER")))
	if err != nil {
		return fmt.Errorf("Ошибка подключения к Kafka: %v", err)
	}

	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("Ошибка получения контроллера: %v", err)

	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("подключения к контроллеру: %v", err)

	}

	defer controllerConn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             fmt.Sprintf("%s", os.Getenv("KAFKA_PRODUCER_TOPIC")),
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		return fmt.Errorf("Топик '%s' уже существует или ошибка создания: %v", fmt.Sprintf("%s", os.Getenv("KAFKA_PRODUCER_TOPIC")), err)
	} else {
		log.Printf("Топик '%s' успешно создан", fmt.Sprintf("%s", os.Getenv("KAFKA_PRODUCER_TOPIC")))
		return nil
	}
}

func NewBookingEventsProducer() BookingEventsProducer {
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(fmt.Sprintf("%s", os.Getenv("KAFKA_BROKER"))),
		Topic:    fmt.Sprintf("%s", os.Getenv("KAFKA_PRODUCER_TOPIC")),
		Balancer: &kafka.LeastBytes{},
	}
	log.Println("Kafka Writer инициализирован")

	return BookingEventsProducer{
		writer: kafkaWriter,
	}
}

func InitKafka() error {
	return createTopic()
}
