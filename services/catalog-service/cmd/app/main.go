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

	services := broker.NewProducer(kafkaCfg, "services")
	defer services.Close()

	specialist := broker.NewProducer(kafkaCfg, "specialist")
	defer specialist.Close()

	schedules := broker.NewProducer(kafkaCfg, "schedules")
	defer schedules.Close()

	serviceRepo := repository.NewServicesRepositry(db)
	specialistRepo := repository.NewSpecialistRepository(db)
	scheduleRepo := repository.NewSchedulesRepository(db)

	serivceService := service.NewServicesService(specialistRepo, serviceRepo, services)
	specialistService := service.NewSpecialistService(specialistRepo, specialist)
	scheduleService := service.NewSchedulesService(scheduleRepo, schedules)

	router := gin.Default()
	transport.RegisterRoutes(router, serivceService, specialistService, scheduleService)
	if err := router.Run(); err != nil {
		log.Fatalf("не удалось запустить HTTP-сервер: %v", err)
	}
	router.Run(":8080")
}
