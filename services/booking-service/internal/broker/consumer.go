package broker

import (
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
)

type BookingEventsConsumer struct {
	Reader *kafka.Reader
}

func NewBookingEventsConsumer() *BookingEventsConsumer {
	return &BookingEventsConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{fmt.Sprintf("%s", os.Getenv("KAFKA_BROKER"))},
			Topic:    fmt.Sprintf("%s", os.Getenv("KAFKA_CONSUMER_TOPIC")),
			GroupID:  "booking-service",
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
	}
}

func (c BookingEventsConsumer) Close() error {
	if c.Reader == nil {
		return nil
	}

	return c.Reader.Close()
}
