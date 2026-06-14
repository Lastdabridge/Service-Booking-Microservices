package service

import (
	"errors"

	"gorm.io/gorm"
	"github.com/Veoler/notification-audit-service/internal/model"
	"github.com/Veoler/notification-audit-service/internal/repository"
)

var (
    ErrNotificationNotFound = errors.New("notification not found")
    ErrForbidden            = errors.New("access denied")
)

type NotificationService interface {
	CreateNotification(req model.NotificationCreateRequest) (*model.Notification, error)
	GetMyNotifications(userID uint) ([]model.Notification, error)
	MarkNotificationAsRead(notificationID, userID uint) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) CreateNotification(req model.NotificationCreateRequest) (*model.Notification, error) {
	notification := &model.Notification{
		UserID:  req.UserID,
		Type:    req.Type,
		Title:   req.Title,
		Message: req.Message,
		IsRead:  false,
	}

	if err := s.repo.Create(notification); err != nil {
		return nil, err
	}

	return notification, nil
}

func (s *notificationService) GetMyNotifications(userID uint) ([]model.Notification, error) {
	return s.repo.GetByUserID(userID)
}

func (s *notificationService) MarkNotificationAsRead(notificationID, userID uint) error {
	n, err := s.repo.GetByID(notificationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}
		return err
	}

	if n.UserID != userID {
		return ErrForbidden
	}

	return s.repo.MarkAsRead(notificationID)
}
