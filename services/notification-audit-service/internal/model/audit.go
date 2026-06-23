package model

import (
	"time"

	"gorm.io/gorm"
)

type AuditLog struct {
	gorm.Model
	EventType     string    `json:"event_type"     gorm:"type:varchar(128);not null;index"`
	ActorID       uint      `json:"actor_id"       gorm:"index"`
	EntityType    string    `json:"entity_type"    gorm:"type:varchar(64)"`
	EntityID      uint      `json:"entity_id"      gorm:"index"`
	SourceService string    `json:"source_service" gorm:"type:varchar(64)"`
	Payload       string    `json:"payload"        gorm:"type:text"`
}

type AuditLogCreatedRequest struct {
	EventType		string		`json:"event_type"`  
	ActorID			uint		`json:"actor_id"`
	EntityType		string		`json:"entity_type"`
	EntityID		uint		`json:"entity_id"`
	SourceService	string		`json:"source_service"`
	Payload			string  	`json:"payload"`
	CreatedAt     	time.Time	`json:"created_at"` 
}

