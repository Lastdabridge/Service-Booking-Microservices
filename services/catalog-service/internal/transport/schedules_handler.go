package transport

import (
	"catalog-service/internal/dto"
	"catalog-service/internal/middleware"
	"catalog-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SchedulesHandler struct {
	schedule service.SchedulesService
}

func NewSchedulesHandler(schedule service.SchedulesService) *SchedulesHandler {
	return &SchedulesHandler{schedule: schedule}
}

func (h *SchedulesHandler) RegisterRoutes(r *gin.Engine) {
	specialists := r.Group("/specialists/")

	specialists.GET(":id/schedule", h.GetSpecialistSchedule)

	specialists.Use(middleware.RoleMiddleware("admin"))
	{
		specialists.POST(":id/schedule", h.CreateSchedules)
		specialists.PATCH(":id/schedule", h.UpdateSchedules)
		specialists.DELETE(":id/schedule", h.DeleteSchedules)
	}
}

func (h *SchedulesHandler) GetSpecialistSchedule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sched, err := h.schedule.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sched)
}

func (h *SchedulesHandler) CreateSchedules(c *gin.Context) {
	var req dto.ScheduleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sched, err := h.schedule.CreateSchedules(c, uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, sched)
}

func (h *SchedulesHandler) UpdateSchedules(c *gin.Context) {
	var req dto.ScheduleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sched, err := h.schedule.UpdateSchedules(c, uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sched)
}

func (h *SchedulesHandler) DeleteSchedules(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.schedule.DeleteSchedules(c, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "расписание удалено"})
}
