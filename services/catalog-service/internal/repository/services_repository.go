package repository

import (
	"catalog-service/internal/models"

	"gorm.io/gorm"
)

type ServicesRepository interface {
	GetServices() ([]models.Service, error)

	CreateService(req *models.Service) error

	CreateSpecServ(req *models.SpecialistService) error

	UpdateService(id uint, req *models.Service) error

	DeleteService(id uint) error

	DeleteSpecServ(id uint) error

	GetByID(id uint) (*models.Service, error)

	GetByIDSpecServ(id uint) (*models.SpecialistService, error)
}

type gormServicesRepository struct {
	db *gorm.DB
}

func NewServicesRepositry(db *gorm.DB) ServicesRepository {
	return &gormServicesRepository{db: db}
}

func (r *gormServicesRepository) GetServices() ([]models.Service, error) {
	var service []models.Service
	if err := r.db.Find(&service).Error; err != nil {
		return nil, err
	}
	return service, nil
}

func (r *gormServicesRepository) CreateService(req *models.Service) error {
	return r.db.Create(req).Error
}

func (r *gormServicesRepository) CreateSpecServ(req *models.SpecialistService) error {
	return r.db.Create(req).Error
}

func (r *gormServicesRepository) UpdateService(id uint, req *models.Service) error {
	return r.db.Model(&models.Service{}).
		Where("id = ?", id).
		Updates(req).Error
}

func (r *gormServicesRepository) DeleteService(id uint) error {
	return r.db.Delete(&models.Service{}, id).Error
}

func (r *gormServicesRepository) DeleteSpecServ(id uint) error {
	return r.db.Delete(&models.SpecialistService{}, id).Error
}

func (r *gormServicesRepository) GetByID(id uint) (*models.Service, error) {
	var service models.Service
	if err := r.db.First(&service, id).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *gormServicesRepository) GetByIDSpecServ(id uint) (*models.SpecialistService, error) {
	var specServ models.SpecialistService
	if err := r.db.First(&specServ, id).Error; err != nil {
		return nil, err
	}
	return &specServ, nil
}
