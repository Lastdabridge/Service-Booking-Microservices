package broker

import (
	"catalog-service/internal/config"
	"context"

	"github.com/segmentio/kafka-go"
)

type Handler func(ctx context.Context, value []byte) error

type Consumer struct {
	reader  *kafka.Reader
	handler Handler
}

func NewConsumer(cfg config.KafkaConfig, topic string, handler Handler) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   topic, // топик при создании консьюмера
		GroupID: cfg.GroupID,
	})
	return &Consumer{
		reader:  r,
		handler: handler,
	}
}
