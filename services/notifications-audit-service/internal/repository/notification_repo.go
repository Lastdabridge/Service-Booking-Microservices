package repository

import (
	"context"
	"github.com/Veoler/notifications-audit-service/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	GetByUserID(ctx context.Context, userID uint) ([]model.Notification, error)
	GetByID(ctx context.Context, id uint) (*model.Notification, error)
	MarkAsRead(ctx context.Context, id uint) error
}

type notificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(ctx context.Context, notification *model.Notification) error {
	// если ctx отменён, запрос к БД прервётся, а не будет висеть до завершения.
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepo) GetByUserID(ctx context.Context, userID uint) ([]model.Notification, error) {
	var notifications []model.Notification
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).
	Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepo) GetByID(ctx context.Context, id uint) (*model.Notification, error) {
	var notification model.Notification
	if err := r.db.WithContext(ctx).First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepo) MarkAsRead(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).Where("id = ?", id).
	Update("is_read", true).Error
}

