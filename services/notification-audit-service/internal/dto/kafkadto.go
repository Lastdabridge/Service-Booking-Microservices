package kafkadto

import (
	"time"
	"github.com/Veoler/notification-audit-service/internal/model"

)

type NotificationDTO struct {
	Event          string    				`json:"event"`           
	NotificationID uint      				`json:"notification_id"` 
	UserID         uint      				`json:"user_id"`
	Type           model.NotificationType   `json:"type"`            
	SourceEvent    string    				`json:"source_event"`    
	CreatedAt      time.Time 				`json:"created_at"`
}

type AuditLogDTO struct {
	Event         	string    	`json:"event"`          
	AuditID       	uint      	`json:"audit_id"`       
	EventType		string   	`json:"event_type"`   
	SourceService 	string    	`json:"source_service"`
	CreatedAt     	time.Time 	`json:"created_at"`
}