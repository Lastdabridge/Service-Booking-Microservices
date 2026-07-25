package transport

import (
	"catalog-service/internal/dto"
	"catalog-service/internal/middleware"
	"catalog-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SpecialistHandler struct {
	specialist service.SpecialistService
}

func NewSpecialistHandler(specialist service.SpecialistService) *SpecialistHandler {
	return &SpecialistHandler{specialist: specialist}
}

func (h *SpecialistHandler) RegisterRoutes(r *gin.Engine) {
	spec := r.Group("/specialists/")
	spec.GET("", h.GetAllSpecialists)
	spec.GET("/search", h.GetSpecialistByName)

	spec.Use(middleware.RoleMiddleware("admin"))
	{
		spec.POST("", h.CreateSpecialist)
		spec.PATCH(":id", h.UpdateSpecialist)
		spec.DELETE(":id", h.DeleteSpecialist)
	}
}

func (h *SpecialistHandler) GetAllSpecialists(c *gin.Context) {
	spec, err := h.specialist.GetAllSpecialists(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

func (h *SpecialistHandler) GetSpecialistByName(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	specialist, err := h.specialist.GetSpecialistsByName(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, specialist)
}

func (h *SpecialistHandler) CreateSpecialist(c *gin.Context) {
	var req dto.SpecialistCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	spec, err := h.specialist.CreateSpecialist(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, spec)
}

func (h *SpecialistHandler) UpdateSpecialist(c *gin.Context) {
	var req dto.SpecialistUpdateRequest

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	spec, err := h.specialist.UpdateSpecialist(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

func (h *SpecialistHandler) DeleteSpecialist(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.specialist.DeleteSpecialist(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "специалист удален"})
}
