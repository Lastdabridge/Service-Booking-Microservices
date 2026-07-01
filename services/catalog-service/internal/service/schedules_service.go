package service

import (
	"catalog-service/internal/broker"
	"catalog-service/internal/dto"
	"catalog-service/internal/models"
	"catalog-service/internal/repository"
	"catalog-service/internal/validation"
	"context"
)

type SchedulesService interface {
	GetByID(id uint) (*models.SpecialistSchedule, error)

	CreateSchedules(c context.Context, id uint, req dto.ScheduleCreateRequest) (*models.SpecialistSchedule, error)

	UpdateSchedules(c context.Context, id uint, req dto.ScheduleUpdateRequest) (*models.SpecialistSchedule, error)

	DeleteSchedules(c context.Context, id uint) error
}

type schedulesService struct {
	service   repository.SchedulesRepository
	producer  *broker.Producer
	validator validation.Validator
}

func NewSchedulesService(
	service repository.SchedulesRepository,
	producer *broker.Producer,
) SchedulesService {
	return &schedulesService{
		service:   service,
		producer:  producer,
		validator: validation.NewValidator(),
	}
}

func (s *schedulesService) GetByID(id uint) (*models.SpecialistSchedule, error) {
	schedule, err := s.service.GetByID(id)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

func (s *schedulesService) CreateSchedules(c context.Context, id uint, req dto.ScheduleCreateRequest) (*models.SpecialistSchedule, error) {
	if err := s.validator.ValidateScheduleCreate(id, req); err != nil {
		return nil, err
	}

	schedule := &models.SpecialistSchedule{
		SpecialistID: id,
		Weekday:      req.Weekday,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	}
	if err := s.service.CreateSchedules(id, schedule); err != nil {
		return nil, err
	}

	event := &broker.SpecialistSchedules{
		Event:        broker.EventSpecialistScheduleUpdated,
		SpecialistID: id,
		Weekday:      req.Weekday,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return nil, err
	}

	return schedule, nil
}

func (s *schedulesService) UpdateSchedules(c context.Context, id uint, req dto.ScheduleUpdateRequest) (*models.SpecialistSchedule, error) {
	schedule, err := s.service.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Weekday != nil {
		schedule.Weekday = *req.Weekday
	}
	if req.StartTime != nil {
		schedule.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		schedule.EndTime = *req.EndTime
	}

	if err := s.service.UpdateSchedules(id, schedule); err != nil {
		return nil, err
	}
	event := &broker.SpecialistSchedules{
		Event:        broker.EventSpecialistScheduleUpdated,
		SpecialistID: schedule.SpecialistID,
		Weekday:      schedule.Weekday,
		StartTime:    schedule.StartTime,
		EndTime:      schedule.EndTime,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return nil, err
	}

	return schedule, nil
}

func (s *schedulesService) DeleteSchedules(c context.Context, id uint) error {
	if _, err := s.service.GetByID(id); err != nil {
		return err
	}
	if err := s.service.DeleteSchedules(id); err != nil {
		return err
	}

	event := &broker.ScheduleDelete{
		Event:        broker.EventSpecialistScheduleDeleted,
		SpecialistID: id,
	}
	if err := s.producer.Produce(c, event); err != nil {
		return nil
	}

	return nil
}
