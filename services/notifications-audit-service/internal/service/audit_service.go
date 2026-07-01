package service

import (
	"errors"
	"context"

	"gorm.io/gorm"
	"github.com/Veoler/notifications-audit-service/internal/model"
	"github.com/Veoler/notifications-audit-service/internal/repository"
)

var ErrAuditLogNotFound = errors.New("audit log not found")

type AuditService interface {
	CreateAuditLog(ctx context.Context, req model.AuditLogCreatedRequest) (*model.AuditLog, error)
	GetAllAuditLogs(ctx context.Context, ) ([]model.AuditLog, error)
	GetAuditLogByID(ctx context.Context, id uint) (*model.AuditLog, error)
}

type AuditProducer interface {
	PublishAuditLogged(ctx context.Context, auditLog *model.AuditLog, sourceService string)
}

type auditService struct {
	repo repository.AuditRepository
	producer AuditProducer
}

func NewAuditService(repo repository.AuditRepository, producer AuditProducer) AuditService {
	return &auditService{repo: repo, producer: producer}
}

func (s *auditService) CreateAuditLog(ctx context.Context, req model.AuditLogCreatedRequest) (*model.AuditLog, error) {
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

	if err := s.repo.Create(ctx, auditLog); err != nil {
		return nil, err
	}

	s.producer.PublishAuditLogged(ctx, auditLog, req.SourceService)

	return auditLog, nil
}

func (s *auditService) GetAllAuditLogs(ctx context.Context, ) ([]model.AuditLog, error) {
	return s.repo.GetAll(ctx)
}

func (s *auditService) GetAuditLogByID(ctx context.Context, id uint) (*model.AuditLog, error) {
	auditLog, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditLogNotFound
		}
		return nil, err
	}
	return auditLog, nil
}
