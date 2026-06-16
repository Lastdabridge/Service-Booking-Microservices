package service

import (
	"catalog-service/internal/broker"
	"catalog-service/internal/dto"
	models "catalog-service/internal/models"
	"catalog-service/internal/repository"
	"context"
	"time"
)

type SpecialistService interface {
	CreateSpecialist(c context.Context, req dto.SpecialistCreateRequest) (*models.Specialist, error)

	UpdateSpecialist(c context.Context, id uint, req dto.SpecialistUpdateRequest) (*models.Specialist, error)

	DeleteSpecialist(c context.Context, id uint) error

	GetAllSpecilist() ([]models.Specialist, error)
}

type specialistService struct {
	service  repository.SpecialistRepository
	producer *broker.Producer
}

func NewSpecialistService(
	service repository.SpecialistRepository,
	producer *broker.Producer,
) SpecialistService {
	return &specialistService{
		service:  service,
		producer: producer,
	}
}

func (s *specialistService) CreateSpecialist(c context.Context, req dto.SpecialistCreateRequest) (*models.Specialist, error) {
	spec := &models.Specialist{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := s.service.CreateSpecialist(*spec); err != nil {
		return nil, err
	}

	if err := s.producer.Produce(c, spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func (s *specialistService) UpdateSpecialist(c context.Context, id uint, req dto.SpecialistUpdateRequest) (*models.Specialist, error) {
	spec, err := s.service.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		spec.Name = *req.Name
	}
	if req.Description != nil {
		spec.Description = *req.Description
	}
	if req.IsActive != nil {
		spec.IsActive = *req.IsActive
	}

	if err := s.service.UpdateSpecialist(id, *spec); err != nil {
		return nil, err
	}
	if err := s.producer.Produce(c, spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func (s *specialistService) DeleteSpecialist(c context.Context, id uint) error {
	spec, err := s.service.GetByID(id)
	if err != nil {
		return err
	}

	event := broker.SpecialistDeleted{
		ID:        id,
		Name:      spec.Name,
		Timestamp: time.Now().UTC(), //
	}

	if err := s.service.DeleteSpecialist(id); err != nil {
		return err
	}

	if err := s.producer.Produce(c, event); err != nil {
		return err
	}

	return nil
}

func (s *specialistService) GetAllSpecilist() ([]models.Specialist, error) {
	spec, err := s.service.GetAllSpecilist()
	if err != nil {
		return nil, err
	}

	return spec, nil
}
