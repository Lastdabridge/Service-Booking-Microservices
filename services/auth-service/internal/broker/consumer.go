package broker

import kafkaGo "github.com/segmentio/kafka-go"

type Consumer struct {
	reader *kafkaGo.Reader
}

func NewConsumer(
	brokers []string,
	topic string,
	groupID string,
) *Consumer {
	return &Consumer{
		reader: kafkaGo.NewReader(
			kafkaGo.ReaderConfig{
				Brokers: brokers,
				GroupID: groupID,
				Topic:   topic,
			},
		),
	}
}

func (c *Consumer) Reader() *kafkaGo.Reader {
	return c.reader
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
