package service

import (
	"errors"

	"gorm.io/gorm"
	"github.com/Veoler/notification-audit-service/internal/model"
	"github.com/Veoler/notification-audit-service/internal/repository"
)

var ErrAuditLogNotFound = errors.New("audit log not found")

type AuditService interface {
	CreateAuditLog(req model.AuditLogCreatedRequest) (*model.AuditLog, error)
	GetAllAuditLogs() ([]model.AuditLog, error)
	GetAuditLogByID(id uint) (*model.AuditLog, error)
}

type auditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) CreateAuditLog(req model.AuditLogCreatedRequest) (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		Model: gorm.Model{
			CreatedAt: req.CreatedAt, 
		},
		EventType:     req.EventType,
		ActorID:       req.ActorID,
		EntityType:    req.EntityType,
		EntityID:      req.EntityID,
		SourceService: req.SourceService,
		Payload:       req.Payload,
	}

	if err := s.repo.Create(auditLog); err != nil {
		return nil, err
	}

	return auditLog, nil
}

func (s *auditService) GetAllAuditLogs() ([]model.AuditLog, error) {
	return s.repo.GetAll()
}

func (s *auditService) GetAuditLogByID(id uint) (*model.AuditLog, error) {
	auditLog, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditLogNotFound
		}
		return nil, err
	}
	return auditLog, nil
}
