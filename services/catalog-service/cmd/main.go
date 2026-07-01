package main

import (
	"catalog-service/internal/broker"
	"catalog-service/internal/config"
	"catalog-service/internal/models"
	"catalog-service/internal/repository"
	"catalog-service/internal/service"
	"catalog-service/internal/transport"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.SetUpDatabaseConnection()

	if err := db.AutoMigrate(
		&models.Service{},
		&models.CatalogEvent{},
		&models.SpecialistSchedule{},
		&models.SpecialistService{},
		&models.Specialist{},
	); err != nil {
		log.Fatalf("не удалось выполнить миграции: %v", err)
	}

	kafkaCfg := config.NewKafkaConfig()

	produce := broker.NewProducer(kafkaCfg, "catalog.events")
	defer produce.Close()

	serviceRepo := repository.NewServicesRepositry(db)
	specialistRepo := repository.NewSpecialistRepository(db)
	scheduleRepo := repository.NewSchedulesRepository(db)
	catalogEventRepo := repository.NewCatalogEventRepository(db)

	serivceService := service.NewServicesService(specialistRepo, serviceRepo, produce)
	specialistService := service.NewSpecialistService(specialistRepo, produce)
	scheduleService := service.NewSchedulesService(scheduleRepo, produce)

	consumer := broker.NewBookingEventsConsumer()
	defer consumer.Close()

	go consumeBookingEvents(consumer, catalogEventRepo)

	router := gin.Default()
	transport.RegisterRoutes(router, serivceService, specialistService, scheduleService)
	if err := router.Run(":8083"); err != nil {
		log.Fatalf("не удалось запустить HTTP-сервер: %v", err)
	}
}

func consumeBookingEvents(consumer *broker.CatalogEventsConsumer, repo repository.CatalogEventRepository) {
	log.Println("Kafka consumer started for booking events")
	ctx := context.Background()

	for {
		msg, err := consumer.Reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("ошибка чтения Kafka сообщения: %v", err)
			continue
		}

		var payload map[string]any
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			log.Printf("ошибка десериализации события: %v", err)
			_ = consumer.Reader.CommitMessages(ctx, msg)
			continue
		}

		eventType, _ := payload["event"].(string)
		log.Printf("получено событие booking: %s", eventType)

		data, err := json.Marshal(payload)
		if err != nil {
			log.Printf("ошибка маршалинга payload: %v", err)
			continue
		}

		event := &models.CatalogEvent{
			EventType: eventType,
			Payload:   string(data),
			CreatedAt: time.Now(),
		}

		if err := repo.Save(event); err != nil {
			log.Printf("ошибка сохранения события в catalog_events: %v", err)
			continue
		}

		if err := consumer.Reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("ошибка коммита Kafka сообщения: %v", err)
		}
	}

	log.Println("Kafka consumer stopped")
}