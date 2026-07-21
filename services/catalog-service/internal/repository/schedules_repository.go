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

type SchedulesRepository interface {
	CreateSchedules(ctx context.Context, specialistID uint, req *models.SpecialistSchedule) error

	UpdateSchedules(ctx context.Context, specialistID uint, req *models.SpecialistSchedule) error

	DeleteSchedules(ctx context.Context, specialistID uint) error

	GetByID(ctx context.Context, specialistID uint) (*models.SpecialistSchedule, error)
}

type gormSchedulesRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewSchedulesRepository(db *gorm.DB, rdb *redis.Client) SchedulesRepository {
	return &gormSchedulesRepository{db: db, rdb: rdb}
}

func (r *gormSchedulesRepository) CreateSchedules(ctx context.Context, specialistID uint, req *models.SpecialistSchedule) error {
	req.SpecialistID = specialistID
	if err := r.db.WithContext(ctx).Create(req).Error; err != nil {
		return err
	}

	if err := r.rdb.Del(ctx, "schedules:all").Err(); err != nil {
		slog.Error("failed to delete cache on schedules create", "error", err)
	}
	return nil
}

func (r *gormSchedulesRepository) UpdateSchedules(ctx context.Context, specialistID uint, req *models.SpecialistSchedule) error {
	var schedule models.SpecialistSchedule
	if err := r.db.WithContext(ctx).First(&schedule, "specialist_id = ?", specialistID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	if err := r.db.WithContext(ctx).Model(&models.SpecialistSchedule{}).Where("specialist_id = ?", specialistID).Updates(req).Error; err != nil {
		return err
	}

	keysToDelete := []string{
		"schedules:all",
		fmt.Sprintf("schedules:id:%d", specialistID),
	}
	if err := r.rdb.Del(ctx, keysToDelete...).Err(); err != nil {
		slog.Error("failed to delete cache on schedules update", "error", err)
	}

	return nil
}

func (r *gormSchedulesRepository) DeleteSchedules(ctx context.Context, specialistID uint) error {
	if err := r.db.WithContext(ctx).First(&models.SpecialistSchedule{}, "specialist_id = ?", specialistID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	if err := r.db.WithContext(ctx).Where("specialist_id = ?", specialistID).Delete(&models.SpecialistSchedule{}).Error; err != nil {
		return err
	}
	keysToDelete := []string{
		"schedules:all",
		fmt.Sprintf("schedules:id:%d", specialistID),
	}
	if err := r.rdb.Del(ctx, keysToDelete...).Err(); err != nil {
		slog.Error("failed to delete cache on schedules delete", "error", err)
	}

	return nil
}

func (r *gormSchedulesRepository) GetByID(ctx context.Context, specialistID uint) (*models.SpecialistSchedule, error) {
	key := fmt.Sprintf("schedules:id:%d", specialistID)
	cached, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var schedule models.SpecialistSchedule
		if err := json.Unmarshal([]byte(cached), &schedule); err == nil {
			return &schedule, nil
		}
		slog.Warn("failed to unmarshal cached schedule by id, falling back to DB", "key", key, "error", err)
	}

	var schedule models.SpecialistSchedule
	if err := r.db.WithContext(ctx).Where("specialist_id = ?", specialistID).First(&schedule).Error; err != nil {
		return nil, err
	}

	data, err := json.Marshal(schedule)
	if err != nil {
		slog.Error("failed to marshal data for cache", "error", err)
		return &schedule, nil
	}
	if err := r.rdb.Set(ctx, key, data, 5*time.Minute).Err(); err != nil {
		slog.Error("failed to set redis cache", "error", err, "cache_key", key)
	}

	return &schedule, nil
}
