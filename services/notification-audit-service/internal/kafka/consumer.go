package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Veoler/notification-audit-service/internal/dto"
	"github.com/Veoler/notification-audit-service/internal/model"
	"github.com/Veoler/notification-audit-service/internal/service"
	"github.com/segmentio/kafka-go"
)

const (
	TopicUsersEvents         = "users.events"
	TopicCatalogEvents       = "catalog.events"
	TopicBookingEvents       = "booking.events"
	TopicGatewayEvents       = "gateway.events"
)

const eventTypeSuspiciousDetected = "suspicious.activity.detected"

func StartConsumers(
	ctx context.Context,
	broker string,
	groupID string,
	notifSvc service.NotificationService,
	auditSvc service.AuditService,
) {
	topics := []string{
		TopicUsersEvents,
		TopicCatalogEvents,
		TopicBookingEvents,
		TopicNotificationsEvents,
		TopicGatewayEvents,
	}

	for _, topic := range topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{broker},
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 1,
			MaxBytes: 10e6,
		})

		go func(r *kafka.Reader, t string) {
			defer r.Close()
			log.Printf("[KAFKA CONSUMER] started listening to topic: %s", t)

			for {
				msg, err := r.FetchMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						log.Printf("[KAFKA CONSUMER] stopped reading topic: %s", t)
						return
					}
					log.Printf("[KAFKA CONSUMER] failed to read message from %s: %v", t, err)
					continue
				}

				log.Printf("[KAFKA CONSUMER] received message from %s: %s", t, string(msg.Value))

				var event kafkadto.KafkaEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					log.Printf("[KAFKA CONSUMER] failed to unmarshal event from %s: %v", t, err)
					r.CommitMessages(ctx, msg)
					continue
				}

				handleMessage(ctx, t, event, string(msg.Value), notifSvc, auditSvc)

				if err := r.CommitMessages(ctx, msg); err != nil {
					log.Printf("[KAFKA CONSUMER] failed to commit message from %s: %v", t, err)
				}
			}
		}(reader, topic)
	}

	log.Printf("[KAFKA CONSUMER] successfully listening to topics: %v", topics)
}

func handleMessage(
	ctx context.Context,
	topic string,
	event kafkadto.KafkaEvent,
	rawPayload string,
	notifSvc service.NotificationService,
	auditSvc service.AuditService,
) {
	saveAuditLog(ctx, topic, event, rawPayload, auditSvc)

	if topic == TopicNotificationsEvents {
		log.Printf("[KAFKA CONSUMER] internal event %s received — logging to audit only", event.Event)
		return
	}

	if event.Event == "access.denied" {
		handleAccessDenied(ctx, event, auditSvc)
		return
	}

	maybeCreateNotification(ctx, event, notifSvc)
}

func saveAuditLog(
	ctx context.Context,
	topic string,
	event kafkadto.KafkaEvent,
	rawPayload string,
	auditSvc service.AuditService,
) (*model.AuditLog, error) {
	var actorID uint
	if event.UserID > 0 {
		actorID = event.UserID
	} else if event.ClientID > 0 {
		actorID = event.ClientID
	}

	var entityID uint
	if event.BookingID > 0 {
		entityID = event.BookingID
	} else if event.ServiceID > 0 {
		entityID = event.ServiceID
	}

	entry, err := auditSvc.CreateAuditLog(ctx, model.AuditLogCreatedRequest{
		EventType:     event.Event,
		EntityType:    entityTypeFromEvent(event.Event),
		SourceService: sourceServiceFromTopic(topic),
		Payload:       rawPayload,
		ActorID:       actorID,
		EntityID:      entityID,
	})
	if err != nil {
		log.Printf("[KAFKA CONSUMER] failed to create audit log for %s: %v", event.Event, err)
		return nil, err
	}

	log.Printf("[KAFKA CONSUMER] audit log saved successfully: event=%s audit_id=%d", event.Event, entry.ID)
	return entry, nil
}

func handleAccessDenied(ctx context.Context, event kafkadto.KafkaEvent, auditSvc service.AuditService) {
	if event.UserID == 0 {
		log.Printf("[SECURITY] access.denied from unknown user: path=%s", event.Path)
		PublishSuspiciousActivity(ctx, 0, event.Path, event.Method, "access denied: unknown user")
		return
	}

	isSuspicious, err := auditSvc.IsSuspicious(ctx, event.UserID, "access.denied", SuspiciousWindow, SuspiciousThreshold)
	if err != nil {
		log.Printf("[SECURITY] failed to check suspicious activity for user_id=%d: %v", event.UserID, err)
		return
	}

	if !isSuspicious {
		return
	}

	alreadyAlerted, err := auditSvc.IsSuspicious(ctx, event.UserID, eventTypeSuspiciousDetected, SuspiciousWindow, 1)
	if err != nil {
		log.Printf("[SECURITY] failed to check existing alert for user_id=%d: %v", event.UserID, err)
		return
	}
	if alreadyAlerted {
		log.Printf("[SECURITY] alert for user_id=%d already sent within current window, skipping", event.UserID)
		return
	}

	log.Printf("[SECURITY] SUSPICIOUS ACTIVITY DETECTED: user_id=%d exceeded %d attempts within %v",
		event.UserID, SuspiciousThreshold, SuspiciousWindow)

	PublishSuspiciousActivity(ctx, event.UserID, event.Path, event.Method,
		"multiple access denied attempts detected")

	_, _ = auditSvc.CreateAuditLog(ctx, model.AuditLogCreatedRequest{
		EventType:     eventTypeSuspiciousDetected,
		EntityType:    "security",
		SourceService: "notification-audit-service",
		Payload:       "",
		ActorID:       event.UserID,
	})
}

func maybeCreateNotification(ctx context.Context, event kafkadto.KafkaEvent, notifSvc service.NotificationService) {
	var req model.NotificationCreateRequest

	switch event.Event {
	case "user.registered":
		req = model.NotificationCreateRequest{
			UserID:  event.UserID,
			Type:    model.NotificationTypeWelcome,
			Title:   "Добро пожаловать!",
			Message: "Вы успешно зарегистрировались в системе.",
		}

	case "booking.created":
		req = model.NotificationCreateRequest{
			UserID:  event.ClientID,
			Type:    model.NotificationTypeBookingCreated,
			Title:   "Запись создана",
			Message: "Ваша запись успешно создана.",
		}

	case "booking.cancelled":
		req = model.NotificationCreateRequest{
			UserID:  event.ClientID,
			Type:    model.NotificationTypeBookingCancelled,
			Title:   "Запись отменена",
			Message: "Ваша запись была отменена.",
		}

	case "booking.completed":
		req = model.NotificationCreateRequest{
			UserID:  event.ClientID,
			Type:    model.NotificationTypeBookingCompleted,
			Title:   "Запись завершена",
			Message: "Ваша запись завершена. Спасибо!",
		}

	default:
		return
	}

	if req.UserID == 0 {
		log.Printf("[KAFKA CONSUMER] event %s missing user_id or client_id", event.Event)
		NewProducer().PublishNotificationFailed(ctx, 0, event.Event, "missing user_id or client_id")
		return
	}

	notif, err := notifSvc.CreateNotification(ctx, req, event.Event)
	if err != nil {
		log.Printf("[KAFKA CONSUMER] failed to create notification for event %s: %v", event.Event, err)
		return
	}

	log.Printf("[KAFKA CONSUMER] notification successfully created: user_id=%d event=%s notif_id=%d",
		req.UserID, event.Event, notif.ID)
}

func sourceServiceFromTopic(topic string) string {
	switch topic {
	case TopicUsersEvents:
		return "gateway-auth-service"
	case TopicCatalogEvents:
		return "catalog-service"
	case TopicBookingEvents:
		return "booking-service"
	case TopicNotificationsEvents:
		return "notification-audit-service"
	case TopicGatewayEvents:
		return "gateway-auth-service"
	default:
		return topic
	}
}

func entityTypeFromEvent(event string) string {
	switch {
	case len(event) >= 12 && event[:12] == "notification":
		return "notification"
	case len(event) >= 10 && event[:10] == "specialist":
		return "specialist"
	case len(event) >= 7 && event[:7] == "booking":
		return "booking"
	case len(event) >= 7 && event[:7] == "service":
		return "service"
	case len(event) >= 6 && event[:6] == "access":
		return "security"
	case len(event) >= 4 && event[:4] == "user":
		return "user"
	default:
		return "unknown"
	}
}