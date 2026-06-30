package repository

import (
	"errors"

	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Register(user *model.User) error

	GetByID(id uint) (*model.User, error)

	GetByEmail(email string) (*model.User, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db:db}
}

func (r *gormUserRepository) Register(user *model.User) error {
	if user == nil {
		return errors.New("user is nil")
	}

	return r.db.Create(user).Error
}

func (r *gormUserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User

	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *gormUserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User

	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}