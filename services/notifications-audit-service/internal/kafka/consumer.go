package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Veoler/notifications-audit-service/internal/config"
	"github.com/Veoler/notifications-audit-service/internal/dto"
	"github.com/Veoler/notifications-audit-service/internal/model"
	"github.com/Veoler/notifications-audit-service/internal/service"
	"github.com/segmentio/kafka-go"
)

func StartConsumers(
	ctx context.Context,
	cfg *config.Config,
	notifSvc service.NotificationService,
	auditSvc service.AuditService,
) {
	topics := []string{
		cfg.TopicUsersEvents,
		cfg.TopicCatalogEvents,
		cfg.TopicBookingEvents,
		cfg.TopicNotificationsEvents,
		cfg.TopicGatewayEvents,
	}

	for _, topic := range topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{cfg.KafkaBroker},
			Topic:    topic,
			GroupID:  cfg.KafkaGroupID,
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

				handleMessage(ctx, t, event, string(msg.Value), cfg, notifSvc, auditSvc)

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
	cfg *config.Config,
	notifSvc service.NotificationService,
	auditSvc service.AuditService,
) {
	saveAuditLog(ctx, topic, event, rawPayload, cfg, auditSvc)
	
	if topic == cfg.TopicNotificationsEvents {
		log.Printf("[KAFKA CONSUMER] internal event %s received — logging to audit only", event.Event)
		return
	}

	maybeCreateNotification(ctx, event, cfg, notifSvc)
}

func saveAuditLog(
	ctx context.Context,
	topic string,
	event kafkadto.KafkaEvent,
	rawPayload string,
	cfg *config.Config,
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
		SourceService: sourceServiceFromTopic(topic, cfg),
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

func maybeCreateNotification(ctx context.Context, event kafkadto.KafkaEvent, cfg *config.Config, notifSvc service.NotificationService) {
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
			Message: "Ваша запись отменена.",
		}

	case "booking.completed":
		req = model.NotificationCreateRequest{
			UserID:  event.ClientID,
			Type:    model.NotificationTypeBookingCompleted,
			Title:   "Запись завершена",
			Message: "Ваша запись завершена. Спасибо!",
		}

	case "booking.status_changed":
		notifReq, ok := notificationForStatusChange(event)
		if !ok {
			return
		}
		req = notifReq	

	default:
		return
	}

	if req.UserID == 0 {
		log.Printf("[KAFKA CONSUMER] event %s missing user_id or client_id", event.Event)
		NewProducer(cfg).PublishNotificationFailed(ctx, 0, event.Event, "missing user_id or client_id")
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

func notificationForStatusChange(event kafkadto.KafkaEvent) (model.NotificationCreateRequest, bool) {
	switch event.Status {
	case "completed":
		return model.NotificationCreateRequest{
			UserID:  event.ClientID,
			Type:    model.NotificationTypeBookingCompleted,
			Title:   "Статус записи изменен",
			Message: "Ваша запись была завершена. Спасибо!",
		}, true
 
	case "cancelled":
		return model.NotificationCreateRequest{
			UserID:  event.ClientID,
			Type:    model.NotificationTypeBookingCancelled,
			Title:   "Статус записи изменен",
			Message: "Ваша запись была отменена.",
		}, true
 
	case "confirmed", "created":
    	return model.NotificationCreateRequest{}, false
 
	default:
		log.Printf("[KAFKA CONSUMER] booking.status_changed: unknow status %q, notification don't created", event.Status)
		return model.NotificationCreateRequest{}, false
	}
}

func sourceServiceFromTopic(topic string, cfg *config.Config) string {
	switch topic {
	case cfg.TopicUsersEvents:
		return "gateway-auth-service"
	case cfg.TopicCatalogEvents:
		return "catalog-service"
	case cfg.TopicBookingEvents:
		return "booking-service"
	case cfg.TopicNotificationsEvents:
		return "notification-audit-service"
	case cfg.TopicGatewayEvents:
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