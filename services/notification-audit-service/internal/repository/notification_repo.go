package repository

import (
	"github.com/Veoler/notification-audit-service/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(n *model.Notification) error
	GetByUserID(userID uint) ([]model.Notification, error)
	GetByID(id uint) (*model.Notification, error)
	MarkAsRead(id uint) error
}

type notificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepo) GetByUserID(userID uint) ([]model.Notification, error) {
	var notifications []model.Notification
	if err := r.db.Where("user_id = ?", userID).
	Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepo) GetByID(id uint) (*model.Notification, error) {
	var notification model.Notification
	if err := r.db.First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepo) MarkAsRead(id uint) error {
	return r.db.Model(&model.Notification{}).Where("id = ?", id).
	Update("is_read", true).Error
}

