package repository

import (
	"catalog-service/internal/models"

	"gorm.io/gorm"
)

type CatalogEventRepository interface {
	Save(event *models.CatalogEvent) error
}

type gormCatalogEventRepository struct {
	db *gorm.DB
}

func NewCatalogEventRepository(db *gorm.DB) CatalogEventRepository {
	return &gormCatalogEventRepository{db: db}
}

func (r *gormCatalogEventRepository) Save(event *models.CatalogEvent) error {
	return r.db.Create(event).Error
}
