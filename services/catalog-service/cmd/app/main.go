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
	// producer := Вызов конструктора()

	serivceService := service.NewServicesService(serviceRepo, producer)
	// serivceService := service.NewServicesService(serviceRepo, producer)

	router := gin.Default()
	transport.RegisterRoutes(router, serivceService)
	if err := router.Run(); err != nil {
		log.Fatalf("не удалось запустить HTTP-сервер: %v", err)
	}
	router.Run(":8080")
}
