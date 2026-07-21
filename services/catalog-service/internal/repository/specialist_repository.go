package repository

import (
	"catalog-service/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type SpecialistRepository interface {
	CreateSpecialist(ctx context.Context, req *models.Specialist) error

	UpdateSpecialist(ctx context.Context, id uint, req *models.Specialist) error

	DeleteSpecialist(ctx context.Context, id uint) error

	GetAllSpecialists(ctx context.Context) ([]models.Specialist, error)

	GetSpecialistByName(ctx context.Context, name string) ([]models.Specialist, error)

	GetByID(ctx context.Context, id uint) (*models.Specialist, error)
}

type gormSpecialistRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewSpecialistRepository(db *gorm.DB, rdb *redis.Client) SpecialistRepository {
	return &gormSpecialistRepository{db: db, rdb: rdb}
}

func (r *gormSpecialistRepository) CreateSpecialist(ctx context.Context, req *models.Specialist) error {
	if err := r.db.WithContext(ctx).Create(req).Error; err != nil {
		return err
	}

	if err := r.rdb.Del(ctx, "specialists:all").Err(); err != nil {
		slog.Error("failed to delete cache on specialist create", "error", err)
	}
	return nil
}

func (r *gormSpecialistRepository) UpdateSpecialist(ctx context.Context, id uint, req *models.Specialist) error {
	var specialists models.Specialist
	if err := r.db.WithContext(ctx).First(&specialists, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	if err := r.db.WithContext(ctx).Model(&models.Specialist{}).Where("id = ?", id).Updates(req).Error; err != nil {
		return err
	}

	keysToDelete := []string{
		"specialists:all",
		fmt.Sprintf("specialists:id:%d", id),
		fmt.Sprintf("specialist:name:%s", req.Name),
		fmt.Sprintf("specialist:name:%s", specialists.Name),
	}

	if err := r.rdb.Del(ctx, keysToDelete...).Err(); err != nil {
		slog.Error("failed to delete cache on specialist update", "error", err)
	}

	return nil
}

func (r *gormSpecialistRepository) DeleteSpecialist(ctx context.Context, id uint) error {
	var specialist models.Specialist
	if err := r.db.WithContext(ctx).First(&specialist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	if err := r.db.WithContext(ctx).Delete(&models.Specialist{}, id).Error; err != nil {
		return err
	}

	keysToDelete := []string{
		"specialists:all",
		fmt.Sprintf("specialists:id:%d", id),
		fmt.Sprintf("specialist:name:%s", specialist.Name),
	}

	if err := r.rdb.Del(ctx, keysToDelete...).Err(); err != nil {
		slog.Error("failed to delete cache on specialist delete", "error", err)
	}

	return nil
}

func (r *gormSpecialistRepository) GetAllSpecialists(ctx context.Context) ([]models.Specialist, error) {
	cached, err := r.rdb.Get(ctx, "specialists:all").Result()
	if err == nil {
		var specialist []models.Specialist
		if err := json.Unmarshal([]byte(cached), &specialist); err == nil {
			return specialist, nil
		}
		slog.Warn("failed to unmarshal cached specialists, falling back to DB", "key", "specialists:all", "error", err)
	}

	var specialist []models.Specialist
	if err := r.db.WithContext(ctx).Find(&specialist).Error; err != nil {
		return nil, err
	}

	data, err := json.Marshal(specialist)
	if err != nil {
		slog.Error("failed to marshal data for cache", "error", err)
		return specialist, nil
	}

	if err := r.rdb.Set(ctx, "specialists:all", data, 5*time.Minute).Err(); err != nil {
		slog.Error("failed to set redis cache", "error", err, "cache_key", "specialists:all")
	}
	return specialist, nil
}

func (r *gormSpecialistRepository) GetSpecialistByName(ctx context.Context, name string) ([]models.Specialist, error) {
	key := fmt.Sprintf("specialist:name:%s", name)
	cached, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var specialist []models.Specialist
		if err := json.Unmarshal([]byte(cached), &specialist); err == nil {
			return specialist, nil
		}
		slog.Warn("failed to unmarshal cached specialist by name, falling back to DB", "key", key, "error", err)
	}

	var specialist []models.Specialist
	if err := r.db.WithContext(ctx).Where("name ILIKE ?", "%"+name+"%").Find(&specialist).Error; err != nil {
		return nil, err
	}

	data, err := json.Marshal(specialist)
	if err != nil {
		slog.Error("failed to marshal data for cache", "error", err)
		return specialist, nil
	}
	if err := r.rdb.Set(ctx, key, data, 1*time.Minute).Err(); err != nil {
		slog.Error("failed to set redis cache", "error", err, "cache_key", key)
	}
	return specialist, nil
}

func (r *gormSpecialistRepository) GetByID(ctx context.Context, id uint) (*models.Specialist, error) {
	key := fmt.Sprintf("specialists:id:%d", id)
	cached, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var specialist models.Specialist
		if err := json.Unmarshal([]byte(cached), &specialist); err == nil {
			return &specialist, nil
		}
		slog.Warn("failed to unmarshal cached specialist by id, falling back to DB", "key", key, "error", err)
	}

	var specialist models.Specialist
	if err := r.db.WithContext(ctx).First(&specialist, id).Error; err != nil {
		return nil, err
	}

	data, err := json.Marshal(specialist)
	if err != nil {
		slog.Error("failed to marshal data for cache", "error", err)
		return &specialist, nil
	}

	if err := r.rdb.Set(ctx, key, data, 5*time.Minute).Err(); err != nil {
		slog.Error("failed to set redis cache", "error", err, "cache_key", key)
	}
	return &specialist, nil
}
