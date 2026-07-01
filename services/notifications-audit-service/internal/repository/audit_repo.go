package repository

import (
	"context"
	"time"
	"github.com/Veoler/notifications-audit-service/internal/model"
	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	GetAll(ctx context.Context, ) ([]model.AuditLog, error)
	GetByID(ctx context.Context, id uint) (*model.AuditLog, error)
	CountRecentByActor(ctx context.Context, actorID uint, eventType string, since time.Time) (int64, error)
}

type auditRepo struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepo{db: db}
}

func (r *auditRepo) Create(ctx context.Context, auditLog *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(auditLog).Error
}

func (r *auditRepo) GetAll(ctx context.Context, ) ([]model.AuditLog, error) {
	var auditLogs []model.AuditLog
	if err := r.db.WithContext(ctx).Find(&auditLogs).Error; err != nil {
		return nil, err
	}
	return auditLogs, nil
}

func (r *auditRepo) GetByID(ctx context.Context, id uint) (*model.AuditLog, error) {
	var auditLog model.AuditLog
	if err := r.db.WithContext(ctx).First(&auditLog, id).Error; err != nil {
		return nil, err
	}
	return &auditLog, nil
}

func (r *auditRepo) CountRecentByActor(ctx context.Context, actorID uint, eventType string, since time.Time) (int64, error) {
	var count int64
 
	if err := r.db.WithContext(ctx).
		Model(&model.AuditLog{}).
		Where("actor_id = ? AND event_type = ? AND created_at >= ?", actorID, eventType, since).
		Count(&count).Error; err != nil {
		return 0, err
	}
 
	return count, nil
}
