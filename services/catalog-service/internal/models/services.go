package models

import (
	"time"

	"gorm.io/gorm"
)

type Service struct {
	gorm.Model
	Title           string
	Description     string
	DurationMinutes int
	Price           int
	IsActive        bool
}

type SpecialistService struct {
	ID           uint `gorm:"primaryKey"`
	SpecialistID uint
	ServiceID    uint
	CreatedAt    time.Time
}
