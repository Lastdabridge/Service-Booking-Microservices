package transport

import (
	"booking-service/internal/dto"
	"booking-service/internal/models"
	"booking-service/internal/services"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AppointmentHandler struct {
	appointment services.AppointmentService
}

func NewAppointmentHanlder(appointment services.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{appointment: appointment}
}

func (h *AppointmentHandler) RegisterRoutes(router *gin.Engine) {
	router.Use(AuthorizationMiddleware)
	appointments := router.Group("/appointments/")
	{
		appointments.GET("my", h.GetMy)
		appointments.GET("all", h.GetAll)
		appointments.GET("specialist/:id", h.GetSpecialist)
		appointments.DELETE(":id", h.Delete)
		appointments.PATCH(":id/status", h.Update)
		appointments.POST("", h.Create)
	}

}

func (h *AppointmentHandler) GetMy(ctx *gin.Context) {
	ID, err := strconv.ParseUint(ctx.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := uint(ID)
	appointments, err := h.appointment.GetAllClientAppointments(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, appointments)
}

func (h *AppointmentHandler) Update(ctx *gin.Context) {
	var req struct {
		Status models.Status `json:"status" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "укажите status"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appointment, err := h.appointment.UpdateAppointmentStatus(uint(id), req.Status)
	if err != nil {
		if errors.Is(err, services.ErrAppointmentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, appointment)
}

func (h *AppointmentHandler) GetAll(ctx *gin.Context) {
	appointments, err := h.appointment.GetAllAppointments()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, appointments)
}

func (h *AppointmentHandler) GetSpecialist(ctx *gin.Context) {
	urlID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	specialistID := uint(urlID)

	appointments, err := h.appointment.GetAllSpecialistAppointments(specialistID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, appointments)
}

func (h *AppointmentHandler) Delete(ctx *gin.Context) {
	appointmentID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	headerUserID, err := strconv.Atoi(ctx.GetHeader("X-User-ID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appointment, err := h.appointment.GetAppointmentByID(uint(appointmentID))
	if err != nil {
		if errors.Is(err, services.ErrAppointmentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if appointment.ClientID != uint(headerUserID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "удалить можно только свой appointment"})
		return
	}

	if err := h.appointment.DeleteAppointmentByID(uint(appointmentID)); err != nil {
		if errors.Is(err, services.ErrAppointmentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

}

func (h *AppointmentHandler) Create(ctx *gin.Context) {
	clientID, err := strconv.Atoi(ctx.GetHeader("X-User-ID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req dto.AppointmentCreateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ClientID = uint(clientID)
	appointment, err := h.appointment.CreateAppointment(req)
	if err != nil {
		if errors.Is(err, services.ErrServiceNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrSpecialistNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrAttachedNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, appointment)
}
