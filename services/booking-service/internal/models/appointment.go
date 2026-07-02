package models

import (
	"time"

	"gorm.io/gorm"
)

type Status string

const (
	StatusCreated   Status = "created"
	StatusConfirmed Status = "confirmed"
	StatusCancelled Status = "cancelled"
	StatusCompleted Status = "completed"
)

type Appointment struct {
	gorm.Model
	ClientID     uint       `json:"client_id"`
	Weekday      string     `json:"weekday"`
	SpecialistID uint       `json:"specialist_id"`
	ServiceID    uint       `json:"service_id"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Status       Status     `json:"status"`
}
