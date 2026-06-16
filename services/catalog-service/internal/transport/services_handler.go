package transport

import (
	"catalog-service/internal/dto"
	"catalog-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServicesHandler struct {
	service service.ServicesService
}

func NewServicesHandler(service service.ServicesService) *ServicesHandler {
	return &ServicesHandler{service: service}
}

func (h *ServicesHandler) RegisterRoutes(r *gin.Engine) {
	services := r.Group("/service")
	{
		services.GET("", h.GetAll)
		services.POST("", h.CreateService)
		services.PATCH("/:id", h.UpdateService)
		services.DELETE("/:id", h.DeleteService)
	}
}

func (h *ServicesHandler) GetAll(c *gin.Context) {
	services, err := h.service.GetServices()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, services)
}

func (h *ServicesHandler) CreateService(c *gin.Context) {
	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	services, err := h.service.CreateService(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, services)
}

func (h *ServicesHandler) UpdateService(c *gin.Context) {
	var req dto.UpdateServiceRequest
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	services, err := h.service.UpdateService(c, uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, services)
}

func (h *ServicesHandler) DeleteService(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = h.service.DeleteService(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "удаление прошло успешно"})
}
