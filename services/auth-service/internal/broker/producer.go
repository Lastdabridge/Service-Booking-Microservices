package broker

import (
	"context"
	kafkaGo "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafkaGo.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafkaGo.Writer{
			Addr:                   kafkaGo.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafkaGo.LeastBytes{},
			BatchSize:              1,
		},
	}
}

func (p *Producer) Send(ctx context.Context, value []byte) error {
	return p.writer.WriteMessages(
		ctx,
		kafkaGo.Message{
			Value: value,
		},
	)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
