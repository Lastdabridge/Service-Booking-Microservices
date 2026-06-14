package repository

import (
	"github.com/Veoler/notification-audit-service/internal/model"
	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(log *model.AuditLog) error
	GetAll() ([]model.AuditLog, error)
	GetByID(id uint) (*model.AuditLog, error)
}

type auditRepo struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepo{db: db}
}

func (r *auditRepo) Create(auditLog *model.AuditLog) error {
	return r.db.Create(auditLog).Error
}

func (r *auditRepo) GetAll() ([]model.AuditLog, error) {
	var auditLogs []model.AuditLog
	if err := r.db.Find(&auditLogs).Error; err != nil {
		return nil, err
	}
	return auditLogs, nil
}

func (r *auditRepo) GetByID(id uint) (*model.AuditLog, error) {
	var auditLog model.AuditLog
	if err := r.db.First(&auditLog, id).Error; err != nil {
		return nil, err
	}
	return &auditLog, nil
}
