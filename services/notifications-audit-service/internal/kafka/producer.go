package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
	
	"github.com/Veoler/notifications-audit-service/internal/config"
	"github.com/Veoler/notifications-audit-service/internal/model"
	"github.com/Veoler/notifications-audit-service/internal/dto"
	"github.com/segmentio/kafka-go"
)

func createTopics(broker string, topics []string) {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		log.Printf("[KAFKA PRODUCER] failed to dial broker for topic creation: %v", err)
		return
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Printf("[KAFKA PRODUCER] failed to get controller: %v", err)
		return
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Printf("[KAFKA PRODUCER] failed to dial controller: %v", err)
		return
	}
	defer controllerConn.Close()

	var topicConfigs []kafka.TopicConfig
	for _, topic := range topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	if err := controllerConn.CreateTopics(topicConfigs...); err != nil {
		log.Printf("[KAFKA PRODUCER] topics already exist or failed to create: %v", err)
	} else {
		log.Printf("[KAFKA PRODUCER] topics successfully created: %v", topics)
	}
}

var kafkaWriter *kafka.Writer

func InitWriter(cfg *config.Config) {
	createTopics(cfg.KafkaBroker, []string{cfg.TopicNotificationsEvents, cfg.TopicAuditEvents})
	
	kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBroker),
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
	}

	log.Println("[KAFKA PRODUCER] initialized successfully")
}

func CloseWriter() {
	if kafkaWriter != nil {
		if err := kafkaWriter.Close(); err != nil {
			log.Printf("[KAFKA PRODUCER] error while closing writer: %v", err)
		}
	}
}

type Producer struct{
	cfg *config.Config
}
 
func NewProducer(cfg *config.Config) *Producer {
	return &Producer{cfg: cfg}
}

func (p *Producer) PublishNotificationCreated(ctx context.Context, notif *model.Notification, sourceEvent string) {
	event := kafkadto.NotificationCreatedEvent{
		Event:          "notification.created",
		NotificationID: notif.ID,
		UserID:         notif.UserID,
		Type:           string(notif.Type),
		SourceEvent:    sourceEvent,
		CreatedAt:      time.Now().Format(time.RFC3339),
	}

	if err := publish(ctx, p.cfg.TopicNotificationsEvents, event); err != nil {
		log.Printf("[KAFKA PRODUCER] failed to publish notification.created: %v", err)
	}
}

func (p *Producer) PublishNotificationRead(ctx context.Context, notif *model.Notification) {
	event := kafkadto.NotificationReadEvent{
		Event:          "notification.read",
		NotificationID: notif.ID,
		UserID:         notif.UserID,
		CreatedAt:      time.Now().Format(time.RFC3339),
	}

	if err := publish(ctx, p.cfg.TopicNotificationsEvents, event); err != nil {
		log.Printf("[KAFKA PRODUCER] failed to publish notification.read: %v", err)
	}
}

func (p *Producer) PublishNotificationFailed(ctx context.Context, userID uint, sourceEvent, reason string) {
	event := kafkadto.NotificationFailedEvent{
		Event:       "notification.failed",
		UserID:      userID,
		SourceEvent: sourceEvent,
		Reason:      reason,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	if err := publish(ctx, p.cfg.TopicNotificationsEvents, event); err != nil {
		log.Printf("[KAFKA PRODUCER] failed to publish notification.failed: %v", err)
	}
}

func (p *Producer) PublishAuditLogged(ctx context.Context, auditLog *model.AuditLog, sourceService string) {
	event := kafkadto.AuditLoggedEvent{
		Event:         "audit.logged",
		AuditID:       auditLog.ID,
		SourceEvent:   auditLog.EventType,
		SourceService: sourceService,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	if err := publish(ctx, p.cfg.TopicAuditEvents, event); err != nil {
		log.Printf("[KAFKA PRODUCER] failed to publish audit.logged: %v", err)
	}
}

func publish(ctx context.Context, topic string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Value: data,
	}

	if err := kafkaWriter.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message to topic %s: %w", topic, err)
	}

	log.Printf("[KAFKA PRODUCER] successfully published to %s: %s", topic, string(data))
	return nil
}