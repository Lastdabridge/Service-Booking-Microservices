package service

import (
	"catalog-service/internal/broker"
	"catalog-service/internal/dto"
	"catalog-service/internal/models"
	"catalog-service/internal/repository"
	"catalog-service/internal/validation"
	"context"
)

type SpecialistService interface {
	CreateSpecialist(c context.Context, req dto.SpecialistCreateRequest) (*models.Specialist, error)

	UpdateSpecialist(c context.Context, id uint, req dto.SpecialistUpdateRequest) (*models.Specialist, error)

	DeleteSpecialist(c context.Context, id uint) error

	GetAllSpecialists(c context.Context) ([]models.Specialist, error)

	GetSpecialistsByName(c context.Context, name string) ([]models.Specialist, error)
}

type specialistService struct {
	service   repository.SpecialistRepository
	producer  *broker.Producer
	validator validation.Validator
}

func NewSpecialistService(
	service repository.SpecialistRepository,
	producer *broker.Producer,
) SpecialistService {
	return &specialistService{
		service:   service,
		producer:  producer,
		validator: validation.NewValidator(),
	}
}

func (s *specialistService) CreateSpecialist(c context.Context, req dto.SpecialistCreateRequest) (*models.Specialist, error) {
	if err := s.validator.ValidateSpecialistCreate(req); err != nil {
		return nil, err
	}

	spec := &models.Specialist{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := s.service.CreateSpecialist(c, spec); err != nil {
		return nil, err
	}

	event := &broker.Specialist{
		Event:        broker.EventSpecialistCreated,
		SpecialistID: spec.ID,
		Name:         req.Name,
		Description:  req.Description,
		IsActive:     req.IsActive,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return nil, err
	}

	return spec, nil
}

func (s *specialistService) UpdateSpecialist(c context.Context, id uint, req dto.SpecialistUpdateRequest) (*models.Specialist, error) {
	if err := s.validator.ValidateSpecialistUpdate(req); err != nil {
		return nil, err
	}

	spec, err := s.service.GetByID(c, id)
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

	if err := s.service.UpdateSpecialist(c, id, spec); err != nil {
		return nil, err
	}

	event := &broker.Specialist{
		Event:        broker.EventSpecialistUpdated,
		SpecialistID: spec.ID,
		Name:         spec.Name,
		Description:  spec.Description,
		IsActive:     spec.IsActive,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return nil, err
	}

	return spec, nil
}

func (s *specialistService) DeleteSpecialist(c context.Context, id uint) error {
	_, err := s.service.GetByID(c, id)
	if err != nil {
		return err
	}

	if err := s.service.DeleteSpecialist(c, id); err != nil {
		return err
	}

	event := broker.SpecialistDelete{
		Event: broker.EventSpecialistDeleted,
		ID:    id,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return err
	}

	return nil
}

func (s *specialistService) GetAllSpecialists(c context.Context) ([]models.Specialist, error) {
	spec, err := s.service.GetAllSpecialists(c)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (s *specialistService) GetSpecialistsByName(c context.Context, name string) ([]models.Specialist, error) {
	specialist, err := s.service.GetSpecialistByName(c, name)
	if err != nil {
		return nil, err
	}
	return specialist, nil
}