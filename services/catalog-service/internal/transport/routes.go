package transport

import (
	"catalog-service/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.Engine,
	service service.ServicesService,
) {
	serviceHandler := NewServicesHandler(service)

	serviceHandler.RegisterRoutes(router)
}
