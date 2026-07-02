package repository

import (
	"booking-service/internal/models"
	"errors"

	"gorm.io/gorm"
)

type SpecialistRepository interface {
	Delete(uint) error
	GetByID(id uint) (*models.Specialist, error)
	CheckService(specialistID uint, ServiceID uint) (*models.SpecialistService, error)
	GetSchedule(weekday string, specialist_id uint) (*models.SpecialistShedules, error)
	WithDB(db *gorm.DB) SpecialistRepository
	UpsertSpecialist(event *models.Specialist) error
	UpsertSchedule(event *models.SpecialistShedules) error
	UpsertAttached(*models.SpecialistService) error
	SpecialistServiceDelete(specialistID uint, serviceID uint) error
	SpecialistShedulesDelete(specialistID uint) error
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

func (r *gormSpecialistRepository) SpecialistServiceDelete(specialistID, serviceID uint) error {
	return r.db.Where("specialist_id = ? AND service_id = ?", specialistID, serviceID).Delete(&models.SpecialistService{}).Error
}

func (r *gormSpecialistRepository) SpecialistShedulesDelete(specialistID uint) error {
	return r.db.Where("specialist_id = ?", specialistID).Delete(&models.SpecialistShedules{}).Error
}

func (r *gormSpecialistRepository) UpsertSpecialist(event *models.Specialist) error {
	var existing models.Specialist

	err := r.db.Where("specialist_id = ?", event.SpecialistID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(event).Error
	} else if err != nil {
		return err
	}

	if event.UpdatedAt.After(existing.UpdatedAt) {
		return r.db.Model(&existing).Updates(event).Error
	}

	return nil
}

func (r *gormSpecialistRepository) UpsertSchedule(event *models.SpecialistShedules) error {
	var existing models.SpecialistShedules

	err := r.db.Where("specialist_id = ?", event.SpecialistID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(event).Error
	} else if err != nil {
		return err
	}

	if event.UpdatedAt.After(existing.UpdatedAt) {
		return r.db.Model(&existing).Updates(event).Error
	}

	return nil
}

func (r *gormSpecialistRepository) UpsertAttached(event *models.SpecialistService) error {
	var existing models.SpecialistService

	err := r.db.Where("specialist_id = ? AND service_id = ?", event.SpecialistID, event.ServiceID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(event).Error
	} else if err != nil {
		return err
	}

	if event.UpdatedAt.After(existing.UpdatedAt) {
		return r.db.Model(&existing).Updates(event).Error
	}

	return nil
}

func (r *gormSpecialistRepository) WithDB(db *gorm.DB) SpecialistRepository {
	return &gormSpecialistRepository{db: db}
}

func (r *gormSpecialistRepository) Delete(id uint) error {
	return r.db.Where("specialist_id = ?", id).Delete(&models.Specialist{}).Error
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

func (r *gormSpecialistRepository) GetSchedule(weekday string, specialist_id uint) (*models.SpecialistShedules, error) {
	var schedule models.SpecialistShedules

	if err := r.db.Where("specialist_id = ? AND weekday = ?", specialist_id, weekday).First(&schedule).Error; err != nil {
		return nil, err
	}

	return &schedule, nil
}
