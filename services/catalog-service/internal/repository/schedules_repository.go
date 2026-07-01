package repository

import (
	"catalog-service/internal/models"

	"gorm.io/gorm"
)

type SchedulesRepository interface {
	CreateSchedules(specialistID uint, req *models.SpecialistSchedule) error

	UpdateSchedules(specialistID uint, req *models.SpecialistSchedule) error

	DeleteSchedules(specialistID uint) error

	GetByID(specialistID uint) (*models.SpecialistSchedule, error)
}

type gormSchedulesRepository struct {
	db *gorm.DB
}

func NewSchedulesRepository(db *gorm.DB) SchedulesRepository {
	return &gormSchedulesRepository{db: db}
}

func (r *gormSchedulesRepository) CreateSchedules(specialistID uint, req *models.SpecialistSchedule) error {
	return r.db.Model(&models.SpecialistSchedule{}).
		Where("SpecialistID = ?", specialistID).
		Create(req).Error
}

func (r *gormSchedulesRepository) UpdateSchedules(specialistID uint, req *models.SpecialistSchedule) error {
	return r.db.Model(&models.SpecialistSchedule{}).
		Where("SpecialistID = ?", specialistID).
		Updates(req).Error
}

func (r *gormSchedulesRepository) DeleteSchedules(specialistID uint) error {
	return r.db.Where("specialist_id = ?", specialistID).Delete(&models.SpecialistSchedule{}).Error
}

func (r *gormSchedulesRepository) GetByID(specialistID uint) (*models.SpecialistSchedule, error) {
	var schedule models.SpecialistSchedule
	if err := r.db.Where("specialist_id = ?", specialistID).First(&schedule).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}
