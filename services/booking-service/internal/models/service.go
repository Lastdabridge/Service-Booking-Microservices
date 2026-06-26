// ========================================= KAFKA ======================================= \\
package models

import "gorm.io/gorm"

type Service struct {
	gorm.Model
	Event           *string `json:"event"`
	ServiceID       *uint   `json:"service_id"`
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	DurationMinutes *int    `json:"duration_minutes"`
	Price           *int    `json:"price"`
	IsActive        *bool   `json:"is_active"`
}

type ServiceDelete struct {
	ID uint `json:"id"`
}
