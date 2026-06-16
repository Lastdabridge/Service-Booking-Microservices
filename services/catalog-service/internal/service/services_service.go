package service

import (
	"catalog-service/internal/broker"
	"catalog-service/internal/dto"
	"catalog-service/internal/models"
	"catalog-service/internal/repository"
	"context"
	"errors"
)

type ServicesService interface {
	GetServices() ([]models.Service, error)

	CreateService(c context.Context, req dto.CreateServiceRequest) (*models.Service, error)

	UpdateService(c context.Context, id uint, req dto.UpdateServiceRequest) (*models.Service, error)

	DeleteService(id uint) error
}

type servicesServise struct {
	service  repository.ServicesRepository
	producer *broker.Producer
}

// func NewServicesService(service repository.ServicesRepository, producer broker.AdminActionEventProducer) ServicesService {
func NewServicesService(service repository.ServicesRepository, producer *broker.Producer) ServicesService {
	return &servicesServise{
		service:  service,
		producer: producer,
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

	err := s.producer.Produce(c, service)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (s *servicesServise) UpdateService(c context.Context, id uint, req dto.UpdateServiceRequest) (*models.Service, error) {
	service, err := s.service.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		service.Title = *req.Title
	}
	if req.Description != nil {
		service.Description = *req.Description
	}
	if req.DurationMinutes != nil {
		service.DurationMinutes = *req.DurationMinutes
	}
	if req.Price != nil {
		service.Price = *req.Price
	}
	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}

	if err := s.service.UpdateService(id, service); err != nil {
		return nil, err
	}

	if err := s.producer.Produce(context.Background(), service); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *servicesServise) DeleteService(id uint) error {
	_, err := s.service.GetByID(id)
	if err != nil {
		return errors.New("услуги с таким id не существует")
	}
	if err := s.service.DeleteService(id); err != nil {
		return err
	}

	return nil
}
