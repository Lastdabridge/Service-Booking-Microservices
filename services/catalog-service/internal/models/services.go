package models

import (
	"gorm.io/gorm"
)

type Service struct {
	gorm.Model
	Title           string `gorm:"type:varchar(120);not null;size:120" json:"title" binding:"required,min=2,max=120"`
	Description     string `gorm:"type:text" json:"description" binding:"required,max=2000"`
	DurationMinutes int    `gorm:"not null;check:duration_minutes > 0" json:"duration_minutes" binding:"required,gt=0,lte=1440"`
	Price           int    `gorm:"not null;check:price >= 0" json:"price" binding:"required,gte=0"`
	IsActive        bool   `gorm:"default:true" json:"is_active"`
}

type SpecialistService struct {
	gorm.Model
	SpecialistID uint `gorm:"not null;uniqueIndex:idx_specialist_service"`
	ServiceID    uint `gorm:"not null;uniqueIndex:idx_specialist_service"`
}
