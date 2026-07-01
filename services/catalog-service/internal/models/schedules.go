package models

import (
	"time"

	"gorm.io/gorm"
)

type SpecialistSchedule struct {
	gorm.Model
	SpecialistID uint      `gorm:"not null;index"`
	Weekday      string    `gorm:"not null"`
	StartTime    time.Time `gorm:"not null"`
	EndTime      time.Time `gorm:"not null"`
}
