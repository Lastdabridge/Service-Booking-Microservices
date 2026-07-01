package broker

import (
	"time"

	"gorm.io/gorm"
)

const (
	EventSpecialistCreated         = "specialist.created"
	EventSpecialistUpdated         = "specialist.updated"
	EventSpecialistDeleted         = "specialist.deleted"
	EventSpecialistServiceAttached = "specialist.service_attached"
	EventSpecialistScheduleUpdated = "specialist.schedule_updated"
	EventSpecialistScheduleDeleted = "specialist.schedule_deleted"
	EventSpecialistServiceDeleted  = "specialist.service_deleted"

	EventServiceCreated = "service.created"
	EventServiceUpdated = "service.updated"
	EventServiceDeleted = "service.deleted"
)

// Specialist — для "specialist.created" и "specialist.updated"
type Specialist struct {
	gorm.Model
	Event        string `json:"event"`
	SpecialistID uint   `json:"specialist_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsActive     bool   `json:"is_active"`
}

// SpecialistDelete — для "specialist.deleted"
type SpecialistDelete struct {
	Event string `json:"event"`
	ID    uint   `json:"id"`
}

// SpecialistService — для "specialist.service_attached"
type SpecialistService struct {
	gorm.Model
	Event        string `json:"event"`
	SpecialistID uint   `json:"specialist_id"`
	ServiceID    uint   `json:"service_id"`
}

// SpecialistSchedules — для "specialist.schedule_updated"
type SpecialistSchedules struct {
	gorm.Model
	Event        string    `json:"event"`
	SpecialistID uint      `json:"specialist_id"`
	Weekday      string    `json:"weekday"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

// Service — для "service.created" и "service.updated"
type Service struct {
	gorm.Model
	Event           string `json:"event"`
	ServiceID       *uint  `json:"service_id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	DurationMinutes *int   `json:"duration_minutes"`
	Price           *int   `json:"price"`
	IsActive        *bool  `json:"is_active"`
}

// ServiceDelete — для "service.deleted"
type ServiceDelete struct {
	Event string `json:"event"`
	ID    uint   `json:"id"`
}

type SpecialistServiceDelete struct {
	Event        string `json:"event"`
	SpecialistID uint   `json:"specialist_id"`
	ServiceID    uint   `json:"service_id"`
}

type ScheduleDelete struct {
	Event        string `json:"event"`
	SpecialistID uint   `json:"specialist_id"`
}
