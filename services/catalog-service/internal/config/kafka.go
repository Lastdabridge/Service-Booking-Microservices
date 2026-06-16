package config

import "os"

type KafkaConfig struct {
	Brokers []string
	GroupID string // только для consumer'ов этого сервиса
}

func NewKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKERS")},
		GroupID: os.Getenv("KAFKA_GROUP_ID"),
	}
}
