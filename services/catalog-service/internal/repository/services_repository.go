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

type ServicesRepository interface {
	GetServices(ctx context.Context) ([]models.Service, error)

	GetServiceByTitle(ctx context.Context, title string) ([]models.Service, error)

	CreateService(ctx context.Context, req *models.Service) error

	CreateSpecServ(ctx context.Context, req *models.SpecialistService) error

	UpdateService(ctx context.Context, id uint, req *models.Service) error

	DeleteService(ctx context.Context, id uint) error

	DeleteSpecServ(ctx context.Context, id uint) error

	GetByID(ctx context.Context, id uint) (*models.Service, error)

	GetByIDSpecServ(ctx context.Context, id uint) (*models.SpecialistService, error)
}

type gormServicesRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewServicesRepository(db *gorm.DB, rdb *redis.Client) ServicesRepository {
	return &gormServicesRepository{db: db, rdb: rdb}
}

func (r *gormServicesRepository) GetServices(ctx context.Context) ([]models.Service, error) {
	cached, err := r.rdb.Get(ctx, "services:all").Result()
	if err == nil {
		var services []models.Service
		if err := json.Unmarshal([]byte(cached), &services); err == nil {
			return services, nil // Возвращаем из кеша только если все прошло гладко.
		}
		slog.Warn("failed to unmarshal cached services, falling back to DB", "key", "services:all", "error", err)
	}

	var service []models.Service
	if err := r.db.WithContext(ctx).Find(&service).Error; err != nil {
		return nil, err
	}

	data, err := json.Marshal(service)
	if err != nil {
		slog.Error("failed to marshal data for cache", "error", err)
		return service, nil
	}
	if err := r.rdb.Set(ctx, "services:all", data, 5*time.Minute).Err(); err != nil {
		slog.Error("failed to set redis cache", "error", err, "cache_key", "services:all")
	}
	return service, nil
}

func (r *gormServicesRepository) GetServiceByTitle(ctx context.Context, title string) ([]models.Service, error) {
	key := fmt.Sprintf("service:title:%s", title)
	cached, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var services []models.Service
		if err := json.Unmarshal([]byte(cached), &services); err == nil {
			return services, nil
		}
		slog.Warn("failed to unmarshal cached service by title, falling back to DB", "key", key, "error", err)
	}

	var services []models.Service
	if err := r.db.WithContext(ctx).Where("title ILIKE ?", "%"+title+"%").Find(&services).Error; err != nil {
		return nil, err
	}

	data, err := json.Marshal(services)
	if err != nil {
		slog.Error("failed to marshal data for cache", "error", err)
		return services, nil
	}
	if err := r.rdb.Set(ctx, key, data, 1*time.Minute).Err(); err != nil {
		slog.Error("failed to set redis cache", "error", err, "cache_key", key)
	}
	return services, nil
}

func (r *gormServicesRepository) CreateService(ctx context.Context, req *models.Service) error {
	if err := r.db.WithContext(ctx).Create(req).Error; err != nil {
		return err
	}
	if err := r.rdb.Del(ctx, "services:all").Err(); err != nil {
		slog.Error("failed to delete cache on service create", "error", err)
	}
	return nil
}

func (r *gormServicesRepository) CreateSpecServ(ctx context.Context, req *models.SpecialistService) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *gormServicesRepository) UpdateService(ctx context.Context, id uint, req *models.Service) error {
	var service models.Service
	if err := r.db.WithContext(ctx).First(&service, id).Error; err != nil {
		if gorm.ErrRecordNotFound == err {
			return nil
		}
		return err
	}

	if err := r.db.WithContext(ctx).Model(&models.Service{}).
		Where("id = ?", id).
		Updates(req).Error; err != nil {
		return err
	}
	keysToDelete := []string{
		"services:all",
		fmt.Sprintf("services:id:%d", id),
		fmt.Sprintf("service:title:%s", service.Title),
		fmt.Sprintf("service:title:%s", req.Title), // req.Title - это новое название
	}
	if err := r.rdb.Del(ctx, keysToDelete...).Err(); err != nil {
		slog.Error("failed to delete cache on service update", "error", err)
	}
	return nil
}

func (r *gormServicesRepository) DeleteService(ctx context.Context, id uint) error {
	var service models.Service
	if err := r.db.WithContext(ctx).First(&service, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	if err := r.db.WithContext(ctx).Delete(&models.Service{}, id).Error; err != nil {
		return err
	}
	keysToDelete := []string{
		"services:all",
		fmt.Sprintf("services:id:%d", id),
		fmt.Sprintf("service:title:%s", service.Title),
	}
	if err := r.rdb.Del(ctx, keysToDelete...).Err(); err != nil {
		slog.Error("failed to delete cache on service delete", "error", err)
	}
	return nil
}

func (r *gormServicesRepository) DeleteSpecServ(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.SpecialistService{}, id).Error; err != nil {
		return err
	}

	return nil
}

func (r *gormServicesRepository) GetByID(ctx context.Context, id uint) (*models.Service, error) {
	key := fmt.Sprintf("services:id:%d", id)
	cached, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var services models.Service
		if err := json.Unmarshal([]byte(cached), &services); err == nil {
			return &services, nil
		}
		slog.Warn("failed to unmarshal cached service by id, falling back to DB", "key", key, "error", err)
	}

	var service models.Service
	if err := r.db.WithContext(ctx).First(&service, id).Error; err != nil {
		return nil, err
	}

	data, err := json.Marshal(service)
	if err != nil {
		slog.Error("failed to marshal data for cache", "error", err)
		return &service, nil
	}
	if err := r.rdb.Set(ctx, key, data, 5*time.Minute).Err(); err != nil {
		slog.Error("failed to set redis cache", "error", err, "cache_key", key)
	}
	return &service, nil
}

func (r *gormServicesRepository) GetByIDSpecServ(ctx context.Context, id uint) (*models.SpecialistService, error) {
	var specServ models.SpecialistService
	if err := r.db.WithContext(ctx).First(&specServ, id).Error; err != nil {
		return nil, err
	}
	return &specServ, nil
}