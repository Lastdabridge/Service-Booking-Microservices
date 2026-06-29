package broker

import (
	"context"
	"encoding/json"

	"catalog-service/internal/config"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
}

func NewProducer(cfg config.KafkaConfig, topic string) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  topic,
		Balancer:               &kafka.RoundRobin{},
		AllowAutoTopicCreation: true,
	}
	return &Producer{
		writer: w,
		topic:  topic,
	}
}

func (p *Producer) Produce(ctx context.Context, msg interface{}) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
