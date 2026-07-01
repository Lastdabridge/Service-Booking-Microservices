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
	"fmt"
	"log"
	"time"

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
		log.Printf("Предупреждение: Kafka не запущена: %v", err)
	}

	producer := broker.NewBookingEventsProducer()
	defer producer.Close()

	bookingEventsConsumer := broker.NewBookingEventsConsumer()
	defer bookingEventsConsumer.Close()

	appointmentService := services.NewAppointmentService(appointmentRepo, producer, catalogRepo, specialistRepo, db)
	transport.RegisterRoutes(router, appointmentService)

	go ConsumeEvents(
		bookingEventsConsumer,
		specialistRepo,
		serviceRepo,
	)

	if err := router.Run(":8082"); err != nil {
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

		for {
			var processError error
			switch meta.Event {

			case "specialist.created":
				processError = handleSpecialistCreated(msg.Value, specialistRepo)

			case "specialist.updated":
				processError = handleSpecialistUpdated(msg.Value, specialistRepo)

			case "specialist.deleted":
				processError = handleSpecialistDeleted(msg.Value, specialistRepo)

			case "specialist.service_attached":
				processError = handleSpecialistAttached(msg.Value, specialistRepo)

			case "specialist.service_deleted":
				processError = handleSpecialistServiceDeleted(msg.Value, specialistRepo)

			case "specialist.schedule_updated":
				processError = handleScheduleUpdated(msg.Value, specialistRepo)

			case "specialist.schedule_deleted":
				processError = handleSpecialistScheduleDeleted(msg.Value, specialistRepo)

			case "service.created":
				processError = handleServiceCreated(msg.Value, serviceRepo)

			case "service.updated":
				processError = handleServiceUpdated(msg.Value, serviceRepo)

			case "service.deleted":
				processError = handleServiceDeleted(msg.Value, serviceRepo)

			default:
				log.Printf("неизвестный event: %s", meta.Event)
			}

			if processError == nil {
				break
			}

			log.Printf("ошибка обработки события, повторная попытка через 5 секунд: %v", processError)
			select {
			case <-ctx.Done(): // Если приложение закрывается — мгновенно выходим
				log.Println("Контекст завершен, останавливаем обработку")
				return
			case <-time.After(5 * time.Second): // Иначе просто ждем 5 секунд
			}
		}

		if err := consumer.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("ошибка коммита: %v", err)
		}
	}
	log.Println("Kafka consumer stopped")
}

func handleSpecialistServiceDeleted(
	data []byte,
	repo repository.SpecialistRepository,
) error {
	var event models.SpecialistServiceDeleted

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("критическая ошибка десериализации (сообщение пропущено): %v", err)
		return nil
	}

	if err := repo.SpecialistServiceDelete(event.SpecialistID, event.ServiceID); err != nil {
		return fmt.Errorf("ошибка удаления специалиста с его услугой: %v", err)
	}

	return nil
}

func handleSpecialistScheduleDeleted(
	data []byte,
	repo repository.SpecialistRepository,
) error {
	var event models.SpecialistShedulesDeleted

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("критическая ошибка десериализации (сообщение пропущено): %v", err)
		return nil
	}

	if err := repo.SpecialistShedulesDelete(event.ID); err != nil {
		return fmt.Errorf("ошибка удаления расписания специалиста: %v", err)
	}

	return nil
}

func handleSpecialistCreated(
	data []byte,
	repo repository.SpecialistRepository,
) error {
	var event models.Specialist

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("критическая ошибка десериализации (сообщение пропущено): %v", err)
		return nil
	}

	if err := repo.UpsertSpecialist(&event); err != nil {
		return fmt.Errorf("ошибка создания специалиста: %v", err)
	}

	return nil
}

func handleSpecialistUpdated(
	data []byte,
	repo repository.SpecialistRepository,
) error {
	var event models.Specialist

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("критическая ошибка десериализации (сообщение пропущено): %v", err)
		return nil
	}

	if err := repo.UpsertSpecialist(&event); err != nil {
		return fmt.Errorf("ошибка обновления специалиста: %v", err)
	}

	return nil
}

func handleSpecialistDeleted(
	data []byte,
	repo repository.SpecialistRepository,
) error {
	var event models.SpecialistDelete

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("критическая ошибка десериализации (сообщение пропущено): %v", err)
		return nil
	}

	if err := repo.Delete(event.ID); err != nil {
		if event.ID <= 0 {
			log.Printf("критическая ошибка: невалидный ID услуги (сообщение пропущено): %d", event.ID)
			return nil
		}
	}

	return nil
}

func handleSpecialistAttached(
	data []byte,
	repo repository.SpecialistRepository,
) error {
	var event models.SpecialistService

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("критическая ошибка десериализации (сообщение пропущено): %v", err)
		return nil
	}

	if err := repo.UpsertAttached(&event); err != nil {
		return fmt.Errorf("ошибка привязки услуги: %v", err)
	}
	return nil
}

func handleScheduleUpdated(
	data []byte,
	repo repository.SpecialistRepository,
) error {
	var event models.SpecialistShedules

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("критическая ошибка десериализации (сообщение пропущено): %v", err)
		return nil
	}

	if err := repo.UpsertSchedule(&event); err != nil {
		return fmt.Errorf("ошибка обновления расписания: %v", err)
	}

	return nil
}

func handleServiceCreated(
	data []byte,
	repo repository.ServiceRepository,
) error {
	var event models.Service

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("критическая ошибка десериализации (сообщение пропущено): %v", err)
		return nil
	}

	if err := repo.Upsert(&event); err != nil {
		return fmt.Errorf("ошибка создания услуги: %v", err)
	}

	return nil
}

func handleServiceUpdated(
	data []byte,
	repo repository.ServiceRepository,
) error {
	var event models.Service

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("критическая ошибка десериализации (сообщение пропущено): %v", err)
		return nil
	}

	if err := repo.Upsert(&event); err != nil {
		return fmt.Errorf("ошибка обновления услуги: %v", err)
	}

	return nil
}

func handleServiceDeleted(
	data []byte,
	repo repository.ServiceRepository,
) error {
	var event models.ServiceDelete

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("критическая ошибка десериализации (сообщение пропущено): %v", err)
		return nil
	}

	if event.ID <= 0 {
		log.Printf("критическая ошибка: невалидный ID услуги (сообщение пропущено): %d", event.ID)
		return nil
	}

	if err := repo.Delete(event.ID); err != nil {
		return fmt.Errorf("ошибка удаления услуги: %v", err)
	}

	return nil
}
