package repository

import (
	"booking-service/internal/models"

	"gorm.io/gorm"
)

type SpecialistRepository interface {
	Create(event *models.Specialist) error
	GetLastUpdated() (*models.Specialist, error)
	Delete(uint) error
	GetByID(id uint) (*models.Specialist, error)
	CheckService(specialistID uint, ServiceID uint) (*models.SpecialistService, error)
	CreateAttached(*models.SpecialistService) error
	CreateSchedule(*models.SpecialistShedules) error
	Update(*models.Specialist) error
	GetSchedule(specialist_id uint) (*models.SpecialistShedules, error)
	WithDB(db *gorm.DB) SpecialistRepository
}

type gormSpecialistRepository struct {
	db *gorm.DB
}

func NewSpecialistRepository(
	db *gorm.DB,
) SpecialistRepository {
	return &gormSpecialistRepository{
		db: db,
	}
}

func (r *gormSpecialistRepository) WithDB(db *gorm.DB) SpecialistRepository {
	return &gormSpecialistRepository{db: db}
}

func (r *gormSpecialistRepository) Delete(id uint) error {
	return r.db.Delete(&models.Specialist{}, id).Error
}

func (r *gormSpecialistRepository) Create(event *models.Specialist) error {
	if event == nil {
		return nil
	}

	if err := r.db.Create(event).Error; err != nil {
		return err
	}

	return nil
}

func (r *gormSpecialistRepository) GetLastUpdated() (*models.Specialist, error) {
	var last models.Specialist

	if err := r.db.Where("event = specialist.updated").Order("created_at desc").First(&last).Error; err != nil {
		return nil, err
	}

	return &last, nil
}

func (r *gormSpecialistRepository) GetByID(id uint) (*models.Specialist, error) {
	var specialist models.Specialist

	if err := r.db.Where("specialist_id = ?", id).First(&specialist).Error; err != nil {
		return nil, err
	}
	return &specialist, nil
}

func (r *gormSpecialistRepository) CheckService(specialistID uint, serviceID uint) (*models.SpecialistService, error) {
	var attached models.SpecialistService

	if err := r.db.Where("service_id = ? AND specialist_id = ?", serviceID, specialistID).First(&attached).Error; err != nil {
		return nil, err
	}
	return &attached, nil
}

func (r *gormSpecialistRepository) CreateAttached(req *models.SpecialistService) error {
	if req == nil {
		return nil
	}

	if err := r.db.Create(req).Error; err != nil {
		return err
	}

	return nil
}

func (r *gormSpecialistRepository) CreateSchedule(req *models.SpecialistShedules) error {
	if req == nil {
		return nil
	}
	if err := r.db.Create(req).Error; err != nil {
		return err
	}
	return nil
}

func (r *gormSpecialistRepository) GetSchedule(specialist_id uint) (*models.SpecialistShedules, error) {
	var schedule models.SpecialistShedules

	if err := r.db.Where("specialist_id = ?", specialist_id).First(&schedule).Error; err != nil {
		return nil, err
	}

	return &schedule, nil
}

func (r *gormSpecialistRepository) Update(req *models.Specialist) error {
	if req == nil {
		return nil
	}

	return r.db.Save(req).Error
}
