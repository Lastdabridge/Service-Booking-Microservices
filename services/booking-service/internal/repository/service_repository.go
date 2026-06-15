package repository

import (
	"booking-service/internal/models"

	"gorm.io/gorm"
)

type ServiceRepository interface {
	Create(event *models.Service) error
	GetLastUpdated() (*models.Service, error)
	Delete(uint) error
	GetByID(uint) (*models.Service, error)
	Update(*models.Service) error
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

func (r *gormServiceRepository) WithDB(db *gorm.DB) ServiceRepository {
	return &gormServiceRepository{db: db}
}

func (r *gormServiceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Service{}, id).Error
}

func (r *gormServiceRepository) Create(event *models.Service) error {
	if event == nil {
		return nil
	}

	if err := r.db.Create(event).Error; err != nil {
		return err
	}

	return nil
}

func (r *gormServiceRepository) GetLastUpdated() (*models.Service, error) {
	var last models.Service

	if err := r.db.Where("event = service.updated").Order("created_at desc").First(&last).Error; err != nil {
		return nil, err
	}

	return &last, nil
}

func (r *gormServiceRepository) GetByID(id uint) (*models.Service, error) {
	var service models.Service

	if err := r.db.Where("service_id = ?", id).First(&service).Error; err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *gormServiceRepository) Update(req *models.Service) error {
	if req == nil {
		return nil
	}

	if err := r.db.Save(req).Error; err != nil {
		return err
	}

	return nil
}
