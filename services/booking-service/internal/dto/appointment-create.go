package dto

import (
	"booking-service/internal/models"
	"time"
)

type AppointmentCreateRequest struct {
	ClientID     uint          `json:"client_id"`
	SpecialistID *uint         `json:"specialist_id"`
	Weekday      *string       `json:"weekday"`
	ServiceID    *uint         `json:"service_id"`
	StartTime    *string       `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Status       models.Status `json:"status"`
}
