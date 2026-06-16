package main

import (
	"catalog-service/internal/broker"
	"catalog-service/internal/config"
	"catalog-service/internal/models"
	"catalog-service/internal/repository"
	"catalog-service/internal/service"
	"catalog-service/internal/transport"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.SetUpDatabaseConnection()

	if err := db.AutoMigrate(
		&models.Service{},
	); err != nil {
		log.Fatalf("не удалось выполнить миграции: %v", err)
	}

	kafkaCfg := config.NewKafkaConfig()

	producer := broker.NewProducer(kafkaCfg, "services")
	defer producer.Close()

	serviceRepo := repository.NewServicesRepositry(db)
	specialistRepo := repository.NewSpecialistRepository(db)
	scheduleRepo := repository.NewSchedulesRepository(db)

	serivceService := service.NewServicesService(serviceRepo, producer)
	specialistService := service.NewSpecialistService(specialistRepo, producer)
	scheduleService := service.NewSchedulesService(scheduleRepo, producer)

	router := gin.Default()
	transport.RegisterRoutes(router, serivceService, specialistService, scheduleService)
	if err := router.Run(); err != nil {
		log.Fatalf("не удалось запустить HTTP-сервер: %v", err)
	}
	router.Run(":8080")
}
