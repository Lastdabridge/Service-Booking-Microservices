package service

import (
	"catalog-service/internal/broker"
	"catalog-service/internal/dto"
	"catalog-service/internal/models"
	"catalog-service/internal/repository"
	"catalog-service/internal/validation"
	"context"
)

type ServicesService interface {
	GetServices() ([]models.Service, error)

	CreateService(c context.Context, req dto.CreateServiceRequest) (*models.Service, error)

	CreateSpecServ(c context.Context, req dto.CreateSpecServ) (*models.SpecialistService, error)

	UpdateService(c context.Context, id uint, req dto.UpdateServiceRequest) (*models.Service, error)

	DeleteService(c context.Context, id uint) error

	DeleteSpecServ(c context.Context, id uint) error
}

type servicesServise struct {
	specialist repository.SpecialistRepository
	service    repository.ServicesRepository
	producer   *broker.Producer
	validator  validation.Validator
}

func NewServicesService(
	specialist repository.SpecialistRepository,
	service repository.ServicesRepository,
	producer *broker.Producer,
) ServicesService {
	return &servicesServise{
		specialist: specialist,
		service:    service,
		producer:   producer,
		validator:  validation.NewValidator(),
	}
}

func (s *servicesServise) GetServices() ([]models.Service, error) {
	service, err := s.service.GetServices()
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (s *servicesServise) CreateService(c context.Context, req dto.CreateServiceRequest) (*models.Service, error) {
	if err := s.validator.ValidateServiceCreate(req); err != nil {
		return nil, err
	}

	service := &models.Service{
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
		IsActive:        req.IsActive,
	}

	if err := s.service.CreateService(service); err != nil {
		return nil, err
	}

	event := &broker.Service{
		Event:           broker.EventServiceCreated,
		ServiceID:       &service.ID,
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: &req.DurationMinutes,
		Price:           &req.Price,
		IsActive:        &req.IsActive,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *servicesServise) CreateSpecServ(c context.Context, req dto.CreateSpecServ) (*models.SpecialistService, error) {
	if err := s.validator.ValidateCreateSpecServ(req); err != nil {
		return nil, err
	}

	_, err := s.service.GetByID(req.ServiceID)
	if err != nil {
		return nil, err
	}
	_, err = s.specialist.GetByID(req.SpecialistID)
	if err != nil {
		return nil, err
	}

	specServ := &models.SpecialistService{
		ServiceID:    req.ServiceID,
		SpecialistID: req.SpecialistID,
	}

	if err := s.service.CreateSpecServ(specServ); err != nil {
		return nil, err
	}

	event := &broker.SpecialistService{
		Event:        broker.EventSpecialistServiceAttached,
		SpecialistID: req.SpecialistID,
		ServiceID:    req.ServiceID,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return nil, err
	}

	return specServ, nil
}

func (s *servicesServise) UpdateService(c context.Context, id uint, req dto.UpdateServiceRequest) (*models.Service, error) {
	if err := s.validator.ValidateServiceUpdate(req); err != nil {
		return nil, err
	}

	services, err := s.service.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		services.Title = *req.Title
	}
	if req.Description != nil {
		services.Description = *req.Description
	}
	if req.DurationMinutes != nil {
		services.DurationMinutes = *req.DurationMinutes
	}
	if req.Price != nil {
		services.Price = *req.Price
	}
	if req.IsActive != nil {
		services.IsActive = *req.IsActive
	}

	if err := s.service.UpdateService(id, services); err != nil {
		return nil, err
	}

	event := &broker.Service{
		Event:           broker.EventServiceUpdated,
		ServiceID:       &id,
		Title:           services.Title,
		Description:     services.Description,
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
		IsActive:        req.IsActive,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return nil, err
	}

	return services, nil
}

func (s *servicesServise) DeleteService(c context.Context, id uint) error {
	_, err := s.service.GetByID(id)

	if err != nil {
		return err
	}
	if err := s.service.DeleteService(id); err != nil {
		return err
	}

	event := broker.ServiceDelete{
		Event: broker.EventServiceDeleted,
		ID:    id,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return err
	}
	return nil
}

func (s *servicesServise) DeleteSpecServ(c context.Context, id uint) error {
	specialist, err := s.service.GetByIDSpecServ(id)
	if err != nil {
		return err
	}
	if err := s.service.DeleteSpecServ(id); err != nil {
		return err
	}

	event := broker.SpecialistServiceDelete{
		Event:        broker.EventSpecialistServiceDeleted,
		SpecialistID: specialist.SpecialistID,
		ServiceID:    specialist.ServiceID,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return err
	}

	return nil
}
