package main

import (
	"booking-service/internal/config"
	"booking-service/internal/models"
	"booking-service/internal/repository"
	"booking-service/internal/services"
	"booking-service/internal/transport"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.SetUpDatabaseConnection()

	if err := db.AutoMigrate(&models.Appointment{}); err != nil {
		log.Fatalf("Не удалось мигрировать ошибка: %v", err)
	}

	router := gin.Default()

	appointmentRepo := repository.NewAppointmentRepository(db)
	appointmentService := services.NewAppointmentService(appointmentRepo)
	transport.RegisterRoutes(router, appointmentService)
	if err := router.Run(); err != nil {
		log.Fatalf("не удалось запустить HTTP-сервер: %v", err)
	}
}
