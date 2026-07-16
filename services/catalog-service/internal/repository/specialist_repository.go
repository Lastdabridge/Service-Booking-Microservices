package repository

import (
	"catalog-service/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type SpecialistRepository interface {
	CreateSpecialist(req *models.Specialist) error

	UpdateSpecialist(id uint, req *models.Specialist) error

	DeleteSpecialist(id uint) error

	GetAllSpecilist() ([]models.Specialist, error)

	GetSpecialistByName(name string) ([]models.Specialist, error)

	GetByID(id uint) (*models.Specialist, error)
}

type gormSpecialistRepository struct {
	db *gorm.DB
	rdb *redis.Client
}

func NewSpecialistRepository(db *gorm.DB, rdb *redis.Client) SpecialistRepository {
	return &gormSpecialistRepository{db: db, rdb: rdb}
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
	cached, err := r.rdb.Get(context.Background(), "specialist:all").Result()
	if err == nil {
		var specialist []models.Specialist
		json.Unmarshal([]byte(cached), &specialist)
		return specialist, nil
	}

	var spec []models.Specialist
	if err := r.db.Find(&spec).Error; err != nil {
		return nil, err
	}
	
	data, _ := json.Marshal(spec)
	r.rdb.Set(context.Background(), "specialist:all", data, 5*time.Minute)
	return spec, nil
}

func(r *gormSpecialistRepository) GetSpecialistByName(name string) ([]models.Specialist, error) {
	key := fmt.Sprintf("specialist:title:%s", name)
	cached, err := r.rdb.Get(context.Background(), key).Result()
	if err == nil {
		var specialist []models.Specialist
		json.Unmarshal([]byte(cached), &specialist)
		return specialist, nil
	}

	var specialist []models.Specialist
	if err := r.db.Where("name ILIKE ?", "%"+name+"%").Find(specialist).Error; err != nil {
		return nil, err
	}

	data, _ := json.Marshal(specialist)
	r.rdb.Set(context.Background(), key, data, 5*time.Minute)
	return specialist, nil
}

func (r *gormSpecialistRepository) GetByID(id uint) (*models.Specialist, error) {
	var spec models.Specialist
	if err := r.db.First(&spec, id).Error; err != nil {
		return nil, gorm.ErrRecordNotFound
	}
	return &spec, nil
}
