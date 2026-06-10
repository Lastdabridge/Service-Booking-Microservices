package transport

import (
	"booking-service/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.Engine,
	appointment services.AppointmentService,
) {
	appointmentHandler := NewAppointmentHanlder(appointment)
	appointmentHandler.RegisterRoutes(router)
}
