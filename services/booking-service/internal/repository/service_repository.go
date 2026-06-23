package repository

import (
	"booking-service/internal/models"
	"errors"

	"gorm.io/gorm"
)

type ServiceRepository interface {
	Upsert(service *models.Service) error
	Delete(uint) error
	GetByID(uint) (*models.Service, error)
	WithDB(*gorm.DB) ServiceRepository
}

type gormServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(
	db *gorm.DB,
) ServiceRepository {
	return &gormServiceRepository{
		db: db,
	}
}

func (r *gormServiceRepository) Upsert(event *models.Service) error {
	var existing models.Service

	err := r.db.Where("service_id = ?", event.ServiceID).First(&existing).Error
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

func (r *gormServiceRepository) WithDB(db *gorm.DB) ServiceRepository {
	return &gormServiceRepository{db: db}
}

func (r *gormServiceRepository) Delete(id uint) error {
	return r.db.Where("service_id = ?", id).Delete(&models.Service{}).Error
}

func (r *gormServiceRepository) GetByID(id uint) (*models.Service, error) {
	var service models.Service

	if err := r.db.Where("service_id = ?", id).First(&service).Error; err != nil {
		return nil, err
	}

	return &service, nil
}
