package models

import "gorm.io/gorm"

type Specialist struct {
	gorm.Model
	Name        string `gorm:"type:varchar(100);not null;size:100" json:"name" binding:"required,min=2,max=100"`
	Description string `gorm:"type:text" json:"description" binding:"required,max=1000"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}
