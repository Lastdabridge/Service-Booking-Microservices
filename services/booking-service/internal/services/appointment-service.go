package services

import (
	"booking-service/internal/dto"
	"booking-service/internal/models"
	"booking-service/internal/repository"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrAppointmentNotFound error = errors.New("Appointment по айди не найден")

type AppointmentService interface {
	CreateAppointment(appointment dto.AppointmentCreateRequest) (*models.Appointment, error)
	GetAllClientAppointments(client_id uint) ([]models.Appointment, error)
	GetAllAppointments() ([]models.Appointment, error)
	GetAllSpecialistAppointments(specialist_id uint) ([]models.Appointment, error)
	UpdateAppointmentStatus(appointmentID uint, status models.Status) (*models.Appointment, error)
	DeleteAppointmentByID(appointmentID uint) error
	GetAppointmentByID(appointmentID uint) (*models.Appointment, error)
}

type appointmentService struct {
	appointment repository.AppointmentRepository
}

func NewAppointmentService(
	appointment repository.AppointmentRepository,
) AppointmentService {
	return &appointmentService{
		appointment: appointment,
	}
}

func (s *appointmentService) CreateAppointment(req dto.AppointmentCreateRequest) (*models.Appointment, error) {
	if err := s.isValidCreate(req); err != nil {
		return nil, fmt.Errorf("Ошибка валидации при создании appointmentЖ %v", err)
	}

	timeS := strings.Split(*req.StartTime, ":")
	hour, err := strconv.Atoi(timeS[0])
	if err != nil {
		return nil, errors.New("не удалось преобразовать в int")
	}
	minute, err := strconv.Atoi(timeS[1])
	if err != nil {
		return nil, errors.New("не удалось преобразовать в int")
	}
	if hour < 0 || minute < 0 {
		return nil, errors.New("minute или hour не может быть отрицательным")
	}
	if hour > 23 {
		return nil, errors.New("hour не может быть отрицательным")
	}
	if minute > 59 {
		return nil, errors.New("minute не может быть больше 59")
	}

	var startTime time.Time
	startTime = time.Date(
		startTime.Year(),
		startTime.Month(),
		startTime.Day(),
		hour,
		minute,
		0,
		0,
		startTime.Location(),
	)

	appointment := &models.Appointment{
		ClientID:     req.ClientID,
		SpecialistID: *req.SpecialistID,
		ServiceID:    *req.ServiceID,
		Status:       models.StatusCreated,
		StartTime:    startTime,
		EndTime:      req.EndTime,
	}

	if err := s.appointment.Create(appointment); err != nil {
		return nil, err
	}

	return appointment, nil
}

func (s *appointmentService) isValidCreate(appointment dto.AppointmentCreateRequest) error {
	if appointment.SpecialistID == nil {
		return errors.New("specialist_id обязателен")
	}
	if appointment.ServiceID == nil {
		return errors.New("service_id обязателен")
	}
	if appointment.StartTime == nil {
		return errors.New("start_time обязателен")
	}
	if *appointment.StartTime == "" {
		return errors.New("start_time не должен быть пустым введите в таком формате(час:минута)")
	}
	if !strings.Contains(":", *appointment.StartTime) {
		return errors.New("ваш start_time не содержит \":\"")
	}

	return nil
}

func (s *appointmentService) GetAllClientAppointments(client_id uint) ([]models.Appointment, error) {
	appointments, err := s.appointment.GetAllMy(client_id)
	if err != nil {
		return nil, err
	}

	return appointments, nil
}

func (s *appointmentService) GetAllAppointments() ([]models.Appointment, error) {
	appointments, err := s.appointment.GetAll()
	if err != nil {
		return nil, err
	}

	return appointments, nil
}

func (s *appointmentService) GetAllSpecialistAppointments(specialist_id uint) ([]models.Appointment, error) {
	appointments, err := s.appointment.GetAllSpecialist(specialist_id)
	if err != nil {
		return nil, err
	}

	return appointments, nil
}

func (s *appointmentService) UpdateAppointmentStatus(appointmentID uint, status models.Status) (*models.Appointment, error) {
	appointment, err := s.appointment.GetByID(appointmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}

	if err := s.isValidStatus(status); err != nil {
		return nil, err
	}

	appointment.Status = status

	if err := s.appointment.Update(appointment); err != nil {
		return nil, err
	}

	return appointment, nil
}

func (s *appointmentService) isValidStatus(status models.Status) error {
	switch status {
	case models.StatusCancelled,
		models.StatusCompleted,
		models.StatusConfirmed,
		models.StatusCreated:
		return nil
	}
	return errors.New("такого статуса не существует")
}

func (s *appointmentService) DeleteAppointmentByID(appointmentID uint) error {
	_, err := s.appointment.GetByID(appointmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAppointmentNotFound
		}
		return err
	}

	return s.appointment.Delete(appointmentID)
}

func (s *appointmentService) GetAppointmentByID(appointmentID uint) (*models.Appointment, error) {
	appointment, err := s.appointment.GetByID(appointmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}

	return appointment, nil
}
