package repository

import (
	"booking-service/internal/models"
	"time"

	"gorm.io/gorm"
)

type AppointmentRepository interface {
	Create(appointment *models.Appointment) error
	GetAllMy(client_id uint) ([]models.Appointment, error)
	GetByID(appointment_id uint) (*models.Appointment, error)
	GetAll() ([]models.Appointment, error)
	GetAllSpecialist(specialist_id uint) ([]models.Appointment, error)
	Delete(appointment_id uint) error
	Update(appointment *models.Appointment) error
	HasConflicts(specialistID uint, start time.Time, end time.Time, weekday string) (bool, error)
	WithDB(*gorm.DB) AppointmentRepository
}

type gormAppointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(
	db *gorm.DB,
) AppointmentRepository {
	return &gormAppointmentRepository{
		db: db,
	}
}

func (r *gormAppointmentRepository) WithDB(db *gorm.DB) AppointmentRepository {
	return &gormAppointmentRepository{db: db}
}

func (r *gormAppointmentRepository) Create(appointment *models.Appointment) error {
	if appointment == nil {
		return nil
	}

	if err := r.db.Create(appointment).Error; err != nil {
		return err
	}

	return nil
}

func (r *gormAppointmentRepository) GetAllMy(client_id uint) ([]models.Appointment, error) {
	var appointments []models.Appointment

	if err := r.db.Where("client_id = ?", client_id).Find(&appointments).Error; err != nil {
		return nil, err
	}

	return appointments, nil
}

func (r *gormAppointmentRepository) GetByID(appointment_id uint) (*models.Appointment, error) {
	var appointment models.Appointment

	if err := r.db.First(&appointment, appointment_id).Error; err != nil {
		return nil, err
	}

	return &appointment, nil
}

func (r *gormAppointmentRepository) GetAll() ([]models.Appointment, error) {
	var appointments []models.Appointment

	if err := r.db.Find(&appointments).Error; err != nil {
		return nil, err
	}

	return appointments, nil
}

func (r *gormAppointmentRepository) GetAllSpecialist(specialist_id uint) ([]models.Appointment, error) {
	var appointments []models.Appointment

	if err := r.db.Where("specialist_id = ?", specialist_id).Find(&appointments).Error; err != nil {
		return nil, err
	}

	return appointments, nil
}

func (r *gormAppointmentRepository) Delete(appointment_id uint) error {
	return r.db.Delete(&models.Appointment{}, appointment_id).Error
}

func (r *gormAppointmentRepository) Update(appointment *models.Appointment) error {
	if appointment == nil {
		return nil
	}

	return r.db.Save(appointment).Error
}

func (r *gormAppointmentRepository) HasConflicts(specialistID uint, start time.Time, end time.Time, weekday string) (bool, error) {
	var count int64

	activeStatuses := []models.Status{
		models.StatusCreated,
		models.StatusConfirmed,
	}

	err := r.db.Model(&models.Appointment{}).
		Where("weekday = ?", weekday).
		Where("specialist_id = ?", specialistID).
		Where("status IN ?", activeStatuses).
		Where("start_time < ?", end).
		Where("end_time > ?", start).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil

}
