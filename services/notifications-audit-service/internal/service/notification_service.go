package service

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"github.com/Veoler/notifications-audit-service/internal/model"
	"github.com/Veoler/notifications-audit-service/internal/repository"
)

var (
    ErrNotificationNotFound = errors.New("notification not found")
    ErrForbidden            = errors.New("access denied")
)

type NotificationService interface {
	CreateNotification(ctx context.Context, req model.NotificationCreateRequest, sourceEvent string) (*model.Notification, error)
	GetMyNotifications(ctx context.Context, userID uint) ([]model.Notification, error)
	MarkNotificationAsRead(ctx context.Context, notificationID, userID uint) error
}

type NotificationProducer interface {
	PublishNotificationCreated(ctx context.Context, notif *model.Notification, sourceEvent string)
	PublishNotificationRead(ctx context.Context, notif *model.Notification)
	PublishNotificationFailed(ctx context.Context, userID uint, sourceEvent, reason string)
}

type notificationService struct {
	repo repository.NotificationRepository
	producer NotificationProducer
}

func NewNotificationService(repo repository.NotificationRepository, producer NotificationProducer) NotificationService {
	return &notificationService{repo: repo, producer: producer}
}

func (s *notificationService) CreateNotification(ctx context.Context, req model.NotificationCreateRequest, sourceEvent string) (*model.Notification, error) {
	notification := &model.Notification{
		UserID:  req.UserID,
		Type:    req.Type,
		Title:   req.Title,
		Message: req.Message,
		IsRead:  false,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		s.producer.PublishNotificationFailed(ctx, req.UserID, sourceEvent, err.Error())
		return nil, err
	}

	s.producer.PublishNotificationCreated(ctx, notification, sourceEvent)
	return notification, nil
}

func (s *notificationService) GetMyNotifications(ctx context.Context, userID uint) ([]model.Notification, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *notificationService) MarkNotificationAsRead(ctx context.Context, notificationID, userID uint) error {
	notification, err := s.repo.GetByID(ctx, notificationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}
		return err
	}

	if notification.UserID != userID {
		return ErrForbidden
	}

	if err := s.repo.MarkAsRead(ctx, notificationID); err != nil {
        return err
    }

	
	notification.IsRead = true
	s.producer.PublishNotificationRead(ctx, notification)

    return nil
}
