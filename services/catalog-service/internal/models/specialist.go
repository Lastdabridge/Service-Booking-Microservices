package models

import "gorm.io/gorm"

type Specialist struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	IsActive    bool `gorm:"default:true"`
}
