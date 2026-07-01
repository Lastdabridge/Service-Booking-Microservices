package model

import "gorm.io/gorm"


type NotificationType string

const (
	NotificationTypeWelcome				NotificationType = "welcome"
	NotificationTypeBookingCreated		NotificationType = "booking_created"
	NotificationTypeBookingCancelled	NotificationType = "booking_cancelled"
	NotificationTypeBookingCompleted	NotificationType = "booking_completed"
	NotificationTypeGeneral				NotificationType = "general"
)

type Notification struct {
	gorm.Model
	UserID    uint             `json:"user_id"   gorm:"not null;index"`
	Type      NotificationType `json:"type"      gorm:"type:varchar(64);not null"`
	Title     string           `json:"title"     gorm:"not null"`
	Message   string           `json:"message"   gorm:"type:text;not null"`
	IsRead    bool             `json:"is_read"   gorm:"default:false"`
}

type NotificationCreateRequest struct {
	UserID    uint             `json:"user_id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	IsRead    bool             `json:"is_read"`
}