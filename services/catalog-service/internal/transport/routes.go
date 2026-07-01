package transport

import (
	"catalog-service/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.Engine,
	service service.ServicesService,
	specialist service.SpecialistService,
	schedule service.SchedulesService,
) {
	serviceHandler := NewServicesHandler(service)
	specialistHandler := NewSpecialistHandler(specialist)
	scheduleHandler := NewSchedulesHandler(schedule)

	serviceHandler.RegisterRoutes(router)
	specialistHandler.RegisterRoutes(router)
	scheduleHandler.RegisterRoutes(router)
}
