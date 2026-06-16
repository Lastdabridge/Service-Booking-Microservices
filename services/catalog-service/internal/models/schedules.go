package models

import "gorm.io/gorm"

type SpecialistSchedule struct {
	gorm.Model
	SpecialistID uint   `gorm:"not null;index"`
	Weekday      string `gorm:"not null"`
	StartTime    string `gorm:"not null"`
	EndTime      string `gorm:"not null"`
}
