package main

import (
	"booking-service/internal/broker"
	"booking-service/internal/config"
	"booking-service/internal/models"
	"booking-service/internal/repository"
	"booking-service/internal/services"
	"booking-service/internal/transport"
	"context"
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.SetUpDatabaseConnection()

	if err := db.AutoMigrate(&models.Appointment{}, &models.Service{}, &models.Specialist{}, &models.SpecialistService{}, &models.SpecialistShedules{}); err != nil {
		log.Fatalf("Не удалось мигрировать ошибка: %v", err)
	}

	router := gin.Default()

	appointmentRepo := repository.NewAppointmentRepository(db)
	catalogRepo := repository.NewServiceRepository(db)
	specialistRepo := repository.NewSpecialistRepository(db)
	serviceRepo := repository.NewServiceRepository(db)

	err := broker.InitKafka()
	if err != nil {
		panic(err)
	}

	producer := broker.NewBookingEventsProducer()
	bookingEventsConsumer := broker.NewBookingEventsConsumer()

	appointmentService := services.NewAppointmentService(appointmentRepo, producer, catalogRepo, specialistRepo, db)
	transport.RegisterRoutes(router, appointmentService)

	go ConsumeEvents(
		bookingEventsConsumer,
		specialistRepo,
		serviceRepo,
	)

	if err := router.Run(); err != nil {
		log.Fatalf("не удалось запустить HTTP-сервер: %v", err)
	}

}

type EventMeta struct {
	Event string `json:"event"`
}

func ConsumeEvents(
	consumer *broker.BookingEventsConsumer,
	specialistRepo repository.SpecialistRepository,
	serviceRepo repository.ServiceRepository,
) {
	log.Println("Kafka consumer started")
	defer consumer.Reader.Close()

	ctx := context.Background()

	for {
		msg, err := consumer.Reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}

			log.Printf("ошибка чтения сообщения: %v", err)
			continue
		}

		var meta EventMeta

		if err := json.Unmarshal(msg.Value, &meta); err != nil {
			log.Printf("ошибка десериализации: %v", err)

			_ = consumer.Reader.CommitMessages(ctx, msg)
			continue
		}

		switch meta.Event {

		case "specialist.created":
			handleSpecialistCreated(msg.Value, specialistRepo)

		case "specialist.updated":
			handleSpecialistUpdated(msg.Value, specialistRepo)

		case "specialist.deleted":
			handleSpecialistDeleted(msg.Value, specialistRepo)

		case "specialist.service_attached":
			handleSpecialistAttached(msg.Value, specialistRepo)

		case "specialist.schedule_updated":
			handleScheduleUpdated(msg.Value, specialistRepo)

		case "service.created":
			handleServiceCreated(msg.Value, serviceRepo)

		case "service.updated":
			handleServiceUpdated(msg.Value, serviceRepo)

		case "service.deleted":
			handleServiceDeleted(msg.Value, serviceRepo)

		default:
			log.Printf("неизвестный event: %s", meta.Event)
		}

		if err := consumer.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("ошибка коммита: %v", err)
		}
	}
	log.Println("Kafka consumer stopped")
}

func handleSpecialistCreated(
	data []byte,
	repo repository.SpecialistRepository,
) {
	var event models.Specialist

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("ошибка specialist.created: %v", err)
		return
	}

	if err := repo.Create(&event); err != nil {
		log.Printf("ошибка создания специалиста: %v", err)
	}
}

func handleSpecialistUpdated(
	data []byte,
	repo repository.SpecialistRepository,
) {
	var event models.Specialist

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("ошибка specialist.updated: %v", err)
		return
	}

	if err := repo.Update(&event); err != nil {
		log.Printf("ошибка обновления специалиста: %v", err)
	}
}

func handleSpecialistDeleted(
	data []byte,
	repo repository.SpecialistRepository,
) {
	var event models.Specialist

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("ошибка specialist.deleted: %v", err)
		return
	}

	if err := repo.Delete(event.ID); err != nil {
		log.Printf("ошибка удаления специалиста: %v", err)
	}
}

func handleSpecialistAttached(
	data []byte,
	repo repository.SpecialistRepository,
) {
	var event models.SpecialistService

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("ошибка specialist.service_attached: %v", err)
		return
	}

	if err := repo.CreateAttached(&event); err != nil {
		log.Printf("ошибка привязки услуги: %v", err)
	}
}

func handleScheduleUpdated(
	data []byte,
	repo repository.SpecialistRepository,
) {
	var event models.SpecialistShedules

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("ошибка specialist.schedule_updated: %v", err)
		return
	}

	if err := repo.CreateSchedule(&event); err != nil {
		log.Printf("ошибка обновления расписания: %v", err)
	}
}

func handleServiceCreated(
	data []byte,
	repo repository.ServiceRepository,
) {
	var event models.Service

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("ошибка service.created: %v", err)
		return
	}

	if err := repo.Create(&event); err != nil {
		log.Printf("ошибка создания услуги: %v", err)
	}
}

func handleServiceUpdated(
	data []byte,
	repo repository.ServiceRepository,
) {
	var event models.Service

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("ошибка service.updated: %v", err)
		return
	}

	if err := repo.Update(&event); err != nil {
		log.Printf("ошибка обновления услуги: %v", err)
	}
}

func handleServiceDeleted(
	data []byte,
	repo repository.ServiceRepository,
) {
	var event models.ServiceDelete

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("ошибка service.deleted: %v", err)
		return
	}

	if event.ID <= 0 {
		log.Printf("невалидный ID: %d", event.ID)
		return
	}

	if err := repo.Delete(event.ID); err != nil {
		log.Printf("ошибка удаления услуги: %v", err)
	}
}

/*
func getSpecialistEvent(consumer broker.BookingEventsConsumer, s repository.SpecialistRepository) {
	var event models.Specialist
	defer consumer.Reader.Close()

	ctx := context.Background()
	for {
		msg, err := consumer.Reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("ошибка чтения сообщения: %v", err)
			continue
		}

		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Ошибка десериализации: %v", err)
			consumer.Reader.CommitMessages(ctx, msg)
			continue
		}

		if event.Event == "specialist.created" {

			if err := s.Create(&event); err != nil {
				log.Printf("не удалось создать catalog-event-specialist \"specialist.created\": %v", err)
			}
		} else if event.Event == "specialist.updated" {
			if err := s.Update(&event); err != nil {
				log.Printf("не удалось создать catalog-event-specialist \"specialist.updated\": %v", err)
			}
		} else if event.Event == "specialist.deleted" {
			if event.ID <= 0 {
				log.Println("Невалидный ID при удалении услуги")
			}

			if err := s.Delete(event.ID); err != nil {
				log.Printf("не удалось создать catalog-event-specialist \"specialist.deleted\": %v", err)
			}
		} else {
			log.Print("такого ивента не существует")
		}

		log.Printf("получена модель service: \n%v", event)

		if err := consumer.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("Ошибка коммита: %v", err)
		}

		log.Println("Worker booking-service остановлен")
	}

}

func specialistAttachedAndScheduleUpdatedEvent(consumer broker.BookingEventsConsumer, s repository.SpecialistRepository) {
	var event models.SpecialistService
	var schedule models.SpecialistShedules

	defer consumer.Reader.Close()

	ctx := context.Background()
	for {
		msg, err := consumer.Reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("ошибка чтения сообщения: %v", err)
			continue
		}

		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Ошибка десериализации: %v", err)
			consumer.Reader.CommitMessages(ctx, msg)
			continue
		}

		if event.Event == "specialist.service_attached" {
			if err := s.CreateAttached(&event); err != nil {
				log.Printf("не удалось создать catalog-event-specialist.service_attached \"specialist.service_attached\": %v", err)
			}
		} else if schedule.Event == "specialist.schedule_updated" {
			if err := s.CreateSchedule(&schedule); err != nil {
				log.Printf("не удалось создать catalog-event-specialist.schedule_updated\"specialist.schedule_updated\": %v", err)
			}
		} else {
			log.Print("такого ивента не существует")
		}

		log.Printf("получена модель service_attached: \n%v", event)

		if err := consumer.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("Ошибка коммита: %v", err)
		}

		log.Println("Worker booking-service остановлен")
	}
}

func getServiceEvent(consumer broker.BookingEventsConsumer, s repository.ServiceRepository) {
	var event models.Service

	defer consumer.Reader.Close()

	ctx := context.Background()
	for {
		msg, err := consumer.Reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("ошибка чтения сообщения: %v", err)
			continue
		}

		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Ошибка десериализации: %v", err)
			consumer.Reader.CommitMessages(ctx, msg)
			continue
		}

		if *event.Event == "service.created" {
			if err := s.Create(&event); err != nil {
				log.Printf("не удалось создать catalog-event-service \"service.created\": %v", err)
			}
		} else if *event.Event == "service.updated" {
			if err := s.Update(&event); err != nil {
				log.Printf("не удалось создать catalog-event-service \"service.updated\": %v", err)
			}
		} else if *event.Event == "service.deleted" {
			if event.ID <= 0 {
				log.Println("Невалидный ID при удалении услуги")
			}

			if err := s.Delete(event.ID); err != nil {
				log.Printf("не удалось создать catalog-event-service \"service.deleted\": %v", err)
			}
		} else {
			log.Print("такого ивента не существует")
		}

		log.Printf("получена модель service: \n%v", event)

		if err := consumer.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("Ошибка коммита: %v", err)
		}

		log.Println("Worker booking-service остановлен")
	}
}
*/
