package models

import "time"

type CatalogEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EventType string    `gorm:"type:varchar(100);not null" json:"event_type"`
	Payload   string    `gorm:"type:text;not null" json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}
