// ========================================= KAFKA ======================================= \\
package models

import (
	"time"

	"gorm.io/gorm"
)

type Specialist struct {
	gorm.Model
	Event        string `json:"event"`
	SpecialistID uint   `json:"specialist_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsActive     bool   `json:"is_active"`
}

type SpecialistDelete struct {
	ID uint `json:"id"`
}

// specialist_attached
type SpecialistService struct {
	gorm.Model
	Event        string `json:"event"`
	SpecialistID uint   `json:"specialist_id"`
	ServiceID    uint   `json:"service_id"`
}

type SpecialistServiceDeleted struct {
	SpecialistID uint `json:"specialist_id"`
	ServiceID    uint `json:"service_id"`
}

// specialist_schedules
type SpecialistShedules struct {
	gorm.Model
	Event        string    `json:"event"`
	SpecialistID uint      `json:"specialist_id"`
	Weekday      string    `json:"weekday"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

type SpecialistShedulesDeleted struct {
	ID uint `json:"id"`
}
