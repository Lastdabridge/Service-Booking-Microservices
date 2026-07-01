package broker

import (
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
)

type CatalogEventsConsumer struct {
	Reader *kafka.Reader
}

func NewBookingEventsConsumer() *CatalogEventsConsumer {
	return &CatalogEventsConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{fmt.Sprintf("%s", os.Getenv("KAFKA_BROKERS"))},
			Topic:    fmt.Sprintf("%s", os.Getenv("KAFKA_CONSUMER_TOPIC")),
			GroupID:  "catalog-service",
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
	}
}

func (c CatalogEventsConsumer) Close() error {
	if c.Reader == nil {
		return nil
	}

	return c.Reader.Close()
}
