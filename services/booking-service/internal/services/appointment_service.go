package services

import (
	"booking-service/internal/broker"
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
var ErrServiceNotFound error = errors.New("Service по айди не найден")
var ErrSpecialistNotFound error = errors.New("Specialist по айди не найден")
var ErrAttachedNotFound error = errors.New("такой Specialist Service Attched не найден")

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
	Appointment repository.AppointmentRepository
	Producer    broker.BookingEventsProducer
	Service     repository.ServiceRepository
	Specialist  repository.SpecialistRepository
	db          *gorm.DB
}

func NewAppointmentService(
	appointment repository.AppointmentRepository,
	producer broker.BookingEventsProducer,
	service repository.ServiceRepository,
	specialist repository.SpecialistRepository,
	db *gorm.DB,
) AppointmentService {
	return &appointmentService{
		Appointment: appointment,
		Producer:    producer,
		Service:     service,
		Specialist:  specialist,
		db:          db,
	}
}

func (s *appointmentService) CreateAppointment(req dto.AppointmentCreateRequest) (*models.Appointment, error) {
	var appointment *models.Appointment
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if req.StartTime == nil {
			return errors.New("start_time обязателен")
		}

		startTime, err := s.getStartTime(*req.StartTime)
		if err != nil {
			return err
		}

		if req.ServiceID == nil {
			return errors.New("service_id обязателен")
		}
		endTime, err := s.getEndTime(*req.ServiceID, tx, *startTime)
		if err != nil {
			return err
		}
		if err := s.isValidCreate(tx, req, *startTime, *endTime); err != nil {
			return fmt.Errorf("Ошибка валидации при создании appointment: %v", err)
		}

		if err := s.checkSpecialistWorkDay(tx, *startTime, *endTime, *req.SpecialistID, *req.Weekday); err != nil {
			return err
		}

		appointment = &models.Appointment{
			ClientID:     req.ClientID,
			SpecialistID: *req.SpecialistID,
			ServiceID:    *req.ServiceID,
			Status:       models.StatusCreated,
			StartTime:    startTime,
			EndTime:      endTime,
		}

		if err := s.Appointment.WithDB(tx).Create(appointment); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = s.Producer.PublishBookingEvent(*appointment, "booking.created"); err != nil {
		return nil, err
	}
	return appointment, err
}

func (s *appointmentService) GetAllClientAppointments(client_id uint) ([]models.Appointment, error) {
	appointments, err := s.Appointment.GetAllMy(client_id)
	if err != nil {
		return nil, err
	}

	return appointments, nil
}

func (s *appointmentService) GetAllAppointments() ([]models.Appointment, error) {
	appointments, err := s.Appointment.GetAll()
	if err != nil {
		return nil, err
	}

	return appointments, nil
}

func (s *appointmentService) GetAllSpecialistAppointments(specialist_id uint) ([]models.Appointment, error) {
	appointments, err := s.Appointment.GetAllSpecialist(specialist_id)
	if err != nil {
		return nil, err
	}

	return appointments, nil
}

func (s *appointmentService) UpdateAppointmentStatus(appointmentID uint, status models.Status) (*models.Appointment, error) {
	appointment, err := s.Appointment.GetByID(appointmentID)
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

	if err := s.Appointment.Update(appointment); err != nil {
		return nil, err
	}

	if err = s.Producer.PublishBookingEvent(*appointment, "booking.status_changed"); err != nil {
		return nil, err
	}

	return appointment, nil
}

func (s *appointmentService) DeleteAppointmentByID(appointmentID uint) error {
	appointment, err := s.Appointment.GetByID(appointmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAppointmentNotFound
		}
		return err
	}
	if err := s.Appointment.Delete(appointmentID); err != nil {
		return err
	}

	if err = s.Producer.PublishBookingEvent(*appointment, "booking.cancelled"); err != nil {
		return err
	}

	return nil
}

func (s *appointmentService) GetAppointmentByID(appointmentID uint) (*models.Appointment, error) {
	appointment, err := s.Appointment.GetByID(appointmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}

	return appointment, nil
}

// вспомогательные
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

func (s *appointmentService) getStartTime(startTimeString string) (*time.Time, error) {
	hour, minute, err := s.getHourAndMinute(startTimeString)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	startTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		*hour,
		*minute,
		0,
		0,
		now.Location(),
	)

	if startTime.Before(time.Now()) {
		return nil, errors.New("время не может быть в прошлом")
	}

	return &startTime, nil
}

func (s *appointmentService) getEndTime(serviceID uint, tx *gorm.DB, startTime time.Time) (*time.Time, error) {
	service, err := s.Service.WithDB(tx).GetByID(serviceID)
	if err != nil {
		return nil, err
	}

	if service.DurationMinutes == nil {
		return nil, errors.New("длительность услуги не задана")
	}

	duration := time.Duration(*service.DurationMinutes) * time.Minute
	endTime := startTime.Add(duration)

	return &endTime, nil
}

func (s *appointmentService) getHourAndMinute(timeString string) (hour *int, minute *int, err error) {
	if timeString == "" {
		return nil, nil, errors.New("start_time не должен быть пустым введите в таком формате(час:минута)")
	}
	if !strings.Contains(timeString, ":") {
		return nil, nil, errors.New("ваш start_time не содержит \":\"")
	}
	timeS := strings.Split(timeString, ":")
	if len(timeS) != 2 {
		return nil, nil, errors.New("неверный формат времени, ожидается ЧЧ:ММ")
	}
	hourVal, err := strconv.Atoi(timeS[0])
	if err != nil {
		return nil, nil, errors.New("не удалось преобразовать в int")
	}
	minuteVal, err := strconv.Atoi(timeS[1])
	if err != nil {
		return nil, nil, errors.New("не удалось преобразовать в int")
	}
	if hourVal < 0 || minuteVal < 0 {
		return nil, nil, errors.New("minute или hour не может быть отрицательным")
	}
	if hourVal > 23 {
		return nil, nil, errors.New("hour не может быть больше 23")
	}
	if minuteVal > 59 {
		return nil, nil, errors.New("minute не может быть больше 59")
	}

	return &hourVal, &minuteVal, nil
}

func (s *appointmentService) isValidCreate(tx *gorm.DB, appointment dto.AppointmentCreateRequest, startTime time.Time, endTime time.Time) error {
	if appointment.SpecialistID == nil {
		return errors.New("specialist_id обязателен")
	}
	specialist, err := s.Specialist.WithDB(tx).GetByID(*appointment.SpecialistID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSpecialistNotFound
		}
		return err
	}
	if appointment.ServiceID == nil {
		return errors.New("service_id обязателен")
	}
	if !specialist.IsActive {
		return errors.New("специалист на данный момент не активен")
	}
	service, err := s.Service.WithDB(tx).GetByID(*appointment.ServiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrServiceNotFound
		}
		return err
	}
	if service.IsActive == nil {
		return errors.New("isActive не указан")
	}
	if !*service.IsActive {
		return errors.New("такая услуга не активна")
	}
	attached, err := s.Specialist.WithDB(tx).CheckService(*appointment.SpecialistID, *appointment.ServiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAttachedNotFound
		}
		return err
	}
	if attached.ServiceID != *appointment.ServiceID || attached.SpecialistID != *appointment.SpecialistID {
		return errors.New("специалист не занимается такой услугой")
	}
	if appointment.Weekday == nil {
		return errors.New("weekday обязателен")
	}
	if err := s.isValidWeekDay(*appointment.Weekday); err != nil {
		return err
	}
	hasConflict, err := s.Appointment.WithDB(tx).HasConflicts(*appointment.SpecialistID, startTime, endTime, *appointment.Weekday)
	if err != nil {
		return fmt.Errorf("ошибка проверки конфликтов расписания: %v", err)
	}
	if hasConflict {
		return errors.New("у специалиста уже есть активная запись на это время")
	}

	return nil
}

func (s *appointmentService) isValidWeekDay(weekday string) error {
	switch weekday {
	case "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday":
		return nil
	}
	return errors.New("такой дни недели не существует")
}

func (s *appointmentService) checkSpecialistWorkDay(tx *gorm.DB, startTime, endTime time.Time, specialist_id uint, weekday string) error {
	schedule, err := s.Specialist.WithDB(tx).GetSchedule(weekday, specialist_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("расписание специалиста не найдено")
		}
		return err
	}

	if startTime.Hour() < schedule.StartTime.Hour() ||
		(startTime.Hour() == schedule.StartTime.Hour() && startTime.Minute() < schedule.StartTime.Minute()) {
		return errors.New("специалист в это время еще не работает")
	}
	if endTime.Hour() > schedule.EndTime.Hour() ||
		(endTime.Hour() == schedule.EndTime.Hour() && endTime.Minute() > schedule.EndTime.Minute()) {
		return errors.New("выбранное время не подходит смене специалиста")
	}

	return nil
}
