package main

import (
	"log"

	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/broker"
	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/config"
	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/model"
	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/repository"
	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/service"
	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/transport"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	config.Load()

	db, err := config.SetUpDatabaseConnection()

	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	err = db.AutoMigrate(&model.User{})

	if err != nil {
		log.Fatalf("failed to run model migrations: %v", err)
	}

	conn, controllerConn := broker.NewTopic("users.events")
	defer conn.Close()
	defer controllerConn.Close()

	producer := broker.NewProducer(
		[]string{"kafka:9092"},
		"users.events",
	)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, producer)
	userHandler := transport.NewUserHandler(userService)

	userHandler.RegisterRoutes(r)

	r.Run(":8081")
}
