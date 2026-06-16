// package kafka

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net"
// 	"strconv"

// 	"github.com/segmentio/kafka-go"
// )

// // Producer оборачивает работу с Kafka
// type Producer struct {
// 	writer *kafka.Writer
// 	broker string
// }

// // NewProducer инициализирует продюсер и проверяет/создает нужные топики
// func NewProducer(broker string, topics []string) *Producer {
// 	// Сначала убедимся, что все переданные топики созданы в Kafka
// 	for _, topic := range topics {
// 		ensureTopicExists(broker, topic)
// 	}

// 	// Настраиваем writer без привязки к конкретному топику!
// 	// Если Topic не указан в самом Writer, мы сможем указывать его динамически в каждом сообщении.
// 	w := &kafka.Writer{
// 		Addr:     kafka.TCP(broker),
// 		Balancer: &kafka.LeastBytes{},
// 	}

// 	log.Println("✅ Kafka Producer успешно инициализирован")
// 	return &Producer{
// 		writer: w,
// 		broker: broker,
// 	}
// }

// // SendMessage отправляет любые данные (v) в указанный топик (topic) с ключом (key)
// func (p *Producer) SendMessage(ctx context.Context, topic string, key string, v interface{}) error {
// 	// Сериализуем данные в JSON
// 	payload, err := json.Marshal(v)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal event: %w", err)
// 	}

// 	// Формируем сообщение, явно указывая Target-топик
// 	msg := kafka.Message{
// 		Topic: topic,
// 		Key:   []byte(key),
// 		Value: payload,
// 	}

// 	// Отправляем
// 	if err := p.writer.WriteMessages(ctx, msg); err != nil {
// 		return fmt.Errorf("failed to write message to kafka topic %s: %w", topic, err)
// 	}

// 	log.Printf("📥 Событие успешно отправлено в топик [%s] с ключом %s", topic, key)
// 	return nil
// }

// // Close закрывает коннект к брокеру (вызывается через defer в main)
// func (p *Producer) Close() error {
// 	return p.writer.Close()
// }

// // Внутренняя функция проверки и создания топика
// func ensureTopicExists(broker, topic string) {
// 	conn, err := kafka.Dial("tcp", broker)
// 	if err != nil {
// 		log.Printf("⚠️ Ошибка подключения к Kafka при проверке топика %s: %v", topic, err)
// 		return
// 	}
// 	defer conn.Close()

// 	controller, err := conn.Controller()
// 	if err != nil {
// 		log.Printf("⚠️ Ошибка получения контроллера для %s: %v", topic, err)
// 		return
// 	}

// 	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
// 	if err != nil {
// 		log.Printf("⚠️ Ошибка подключения к контроллеру для %s: %v", topic, err)
// 		return
// 	}
// 	defer controllerConn.Close()

// 	topicConfig := kafka.TopicConfig{
// 		Topic:             topic,
// 		NumPartitions:     1,
// 		ReplicationFactor: 1,
// 	}

// 	err = controllerConn.CreateTopics(topicConfig)
// 	if err != nil {
// 		// Если топик есть — Kafka вернет ошибку, это нормально, игнорируем её
// 		return
// 	}
// 	log.Printf("🆕 Топик '%s' автоматически создан", topic)
// }

// package kafka

// import (
// 	"context"
// 	"encoding/json"
// 	"log"

// 	"github.com/segmentio/kafka-go"
// )

// // NotificationProducer описывает интерфейс для отправки сообщений в Kafka
// type NotificationProducer interface {
// 	Publish(ctx context.Context, topic string, key string, value interface{}) error
// 	Close() error
// }

// type kafkaProducer struct {
// 	writer *kafka.Writer
// }

// // NewNotificationProducer создает и инициализирует новый продюсер Kafka.
// // brokers — это массив адресов (например, []string{"localhost:9092"} или из конфига)
// func NewNotificationProducer(brokers []string) NotificationProducer {
// 	writer := &kafka.Writer{
// 		Addr:     kafka.TCP(brokers...),
// 		Balancer: &kafka.LeastBytes{}, // Балансировщик для эффективного распределения нагрузки
// 		Async:    false,               // Синхронная запись, чтобы мы точно знали, ушло ли сообщение
// 	}

// 	return &kafkaProducer{writer: writer}
// }

// // Publish сериализует структуру в JSON и отправляет её в указанный топик Kafka
// func (p *kafkaProducer) Publish(ctx context.Context, topic string, key string, value interface{}) error {
// 	// Маршалим переданную структуру (например, модель или DTO) в JSON-байты
// 	payload, err := json.Marshal(value)
// 	if err != nil {
// 		log.Printf("failed to marshal kafka payload: %v", err)
// 		return err
// 	}

// 	// Формируем сообщение для Kafka
// 	msg := kafka.Message{
// 		Topic: topic,
// 		Key:   []byte(key),
// 		Value: payload,
// 	}

// 	// Отправляем сообщение в брокер
// 	if err := p.writer.WriteMessages(ctx, msg); err != nil {
// 		log.Printf("failed to write message to kafka topic %s: %v", topic, err)
// 		return err
// 	}

// 	log.Printf("successfully sent message to kafka topic [%s] with key [%s]", topic, key)
// 	return nil
// }

// // Close безопасно закрывает соединение с Kafka при завершении работы приложения
// func (p *kafkaProducer) Close() error {
// 	if err := p.writer.Close(); err != nil {
// 		log.Printf("failed to close kafka writer: %v", err)
// 		return err
// 	}
// 	log.Println("kafka writer closed successfully")
// 	return nil
// }

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/Veoler/notification-audit-service/internal/model"
	"github.com/Veoler/notification-audit-service/internal/dto"
	"github.com/segmentio/kafka-go"
)

const (
	KafkaBroker = "localhost:9092"
	TopicAudit  = "audit.events"
	TopicNotif  = "notifications.events"
)

var KafkaWriter *kafka.Writer

func createTopic() {
	conn, err := kafka.Dial("tcp", KafkaBroker)
	if err != nil {
		log.Printf("Ошибка подключения к Kafka: %v", err)
		return
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Printf("Ошибка получения контроллера: %v", err)
		return
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Printf("Ошибка подключения к контроллеру: %v", err)
		return
	}
	defer controllerConn.Close()

	topicConfig1 := kafka.TopicConfig{
		Topic:             TopicAudit,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	topicConfig2 := kafka.TopicConfig{
		Topic:             TopicNotif,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = controllerConn.CreateTopics(topicConfig1)
	if err != nil {
		log.Printf("Топик '%s' уже существует или ошибка создания: %v", topicConfig1, err)
	} else {
		log.Printf("Топик '%s' успешно создан", topicConfig1)
	}

	err = controllerConn.CreateTopics(topicConfig2)
	if err != nil {
		log.Printf("Топик '%s' уже существует или ошибка создания: %v", topicConfig2, err)
	} else {
		log.Printf("Топик '%s' успешно создан", topicConfig2)
	}
}

func InitKafkaWriter() {
	KafkaWriter = &kafka.Writer{
		Addr: kafka.TCP(KafkaBroker),
		Balancer: &kafka.LeastBytes{},
	}
	log.Println("Kafka Producer (Writer) успешно инициализирован")
}

func PublishAuditLogEvent(action string, auditLog *model.AuditLog) error {
	event := kafkadto.AuditLogDTO{
		Event:         	action,
		AuditID:       	auditLog.ID,
		EventType:   	auditLog.EventType,
		SourceService: 	auditLog.SourceService,
		CreatedAt:     	auditLog.CreatedAt,
	}
	
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("Ошибка сериализации лога в JSON: %v", err)
		return err
	}

	msg := kafka.Message{
		Topic: TopicAudit,
		Key:   []byte(fmt.Sprintf("entity-%d", auditLog.EntityID)), 
		Value: payload,
	}

	err = KafkaWriter.WriteMessages(context.Background(), msg)
	if err != nil {
		log.Printf("Ошибка отправки лога в Kafka топик %s: %v", TopicAudit, err)
		return err
	}

	log.Printf("Событие аудита успешно отправлено в Kafka: ID=%d", auditLog.ID)
	return nil
}

func PublishNotificationEvent(action string, notification *model.Notification) error {
	event := kafkadto.NotificationDTO{
		Event:          action,
		NotificationID: notification.ID,
		UserID:         notification.UserID,
		Type:           notification.Type,
		CreatedAt:      notification.CreatedAt,
	}
	
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("Ошибка сериализации уведомления в JSON: %v", err)
		return err
	}

	msg := kafka.Message{
		Topic: TopicNotif,
		Key:   []byte(fmt.Sprintf("user-%d", notification.UserID)),
		Value: payload,
	}

	err = KafkaWriter.WriteMessages(context.Background(), msg)
	if err != nil {
		log.Printf("Ошибка отправки уведомления в Kafka топик %s: %v", TopicNotif, err)
		return err
	}

	log.Printf("Событие уведомления успешно отправлено в Kafka для UserID=%d", notification.UserID)
	return nil
}

func CloseKafkaWriter() {
	if KafkaWriter != nil {
		if err := KafkaWriter.Close(); err != nil {
			log.Printf("Ошибка при закрытии Kafka Writer: %v", err)
		} else {
			log.Println("Kafka Writer успешно закрыт")
		}
	}
}