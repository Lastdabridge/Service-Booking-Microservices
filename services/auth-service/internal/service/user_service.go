package service

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/apperrors" //
	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/auth"      //
	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/broker"
	model "github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/model"
	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	RegisterUser(req model.UserRegister) error

	LoginUser(req model.UserLogin) (string, error)

	GetMyAccount(id uint) (*model.UserResponse, error)
}

type userService struct {
	users    repository.UserRepository
	producer *broker.Producer
}

func NewUserService(
	users repository.UserRepository,
	producer *broker.Producer,
) UserService {
	return &userService{
		users:    users,
		producer: producer,
	}
}

func (u *userService) RegisterUser(req model.UserRegister) error {
	adminCode := os.Getenv("ADMIN_CODE")

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	role := model.RoleClient

	existingEmail, err := u.users.GetByEmail(email)
	if err != nil {
		return err
	}

	if existingEmail != nil {
		return apperrors.ErrUserAlreadyExists
	}

	if !regexp.MustCompile(`^[\p{L}\s\-']+$`).MatchString(name) {
		return apperrors.ErrNameHasSpecialChars
	}

	if req.AdminCode != nil && *req.AdminCode == adminCode {
		role = model.RoleAdmin
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.ErrPasswordHashFailed
	}

	user := model.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := u.users.Register(&user); err != nil {
		return err
	}

	event := broker.UserEvent{
		Event:     "user.registered",
		UserID:    user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Println("failed to marshal kafka event:", err)
		return err
	}

	err = u.producer.Send(
		context.Background(),
		payload,
	)
	if err != nil {
		log.Println("failed to send event to Kafka:", err)
	}

	return nil
}

func (u *userService) LoginUser(req model.UserLogin) (string, error) {
	user, err := u.users.GetByEmail(req.Email)

	if err != nil {
		return "", err
	}

	if user == nil {
		return "", apperrors.ErrInvalidLoginOrPassword
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	)

	if err != nil {
		return "", apperrors.ErrInvalidLoginOrPassword
	}

	token, err := auth.GenerateToken(user)

	if err != nil {
		return "", err
	}

	event := broker.UserEvent{
		Event:     "user.logged_in",
		UserID:    user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Println("failed to marshal kafka event:", err)
	}

	err = u.producer.Send(
		context.Background(),
		payload,
	)
	if err != nil {
		log.Println("failed to send event to Kafka:", err)
	}

	return token, nil
}

func (u *userService) GetMyAccount(id uint) (*model.UserResponse, error) {
	user, err := u.users.GetByID(id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, apperrors.ErrAccountNotFound
	}

	return &model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
