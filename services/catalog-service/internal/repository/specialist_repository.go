package repository

import (
	"catalog-service/internal/models"

	"gorm.io/gorm"
)

type SpecialistRepository interface {
	CreateSpecialist(req *models.Specialist) error

	UpdateSpecialist(id uint, req *models.Specialist) error

	DeleteSpecialist(id uint) error

	GetAllSpecilist() ([]models.Specialist, error)

	GetByID(id uint) (*models.Specialist, error)
}

type gormSpecialistRepository struct {
	db *gorm.DB
}

func NewSpecialistRepository(db *gorm.DB) SpecialistRepository {
	return &gormSpecialistRepository{db: db}
}

func (r *gormSpecialistRepository) CreateSpecialist(req *models.Specialist) error {
	return r.db.Create(req).Error
}

func (r *gormSpecialistRepository) UpdateSpecialist(id uint, req *models.Specialist) error {
	return r.db.Model(&models.Specialist{}).
		Where("id = ?", id).
		Updates(req).Error
}

func (r *gormSpecialistRepository) DeleteSpecialist(id uint) error {
	return r.db.Delete(&models.Specialist{}, id).Error
}

func (r *gormSpecialistRepository) GetAllSpecilist() ([]models.Specialist, error) {
	var spec []models.Specialist
	if err := r.db.Find(&spec).Error; err != nil {
		return nil, err
	}
	return spec, nil
}

func (r *gormSpecialistRepository) GetByID(id uint) (*models.Specialist, error) {
	var spec models.Specialist
	if err := r.db.First(&spec, id).Error; err != nil {
		return nil, gorm.ErrRecordNotFound
	}
	return &spec, nil
}
