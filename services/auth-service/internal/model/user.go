package model

import (
	"time"
)

type Role string

const (
	RoleClient Role = "client"
	RoleAdmin  Role = "admin"
)

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"type:varchar(100);not null"`
	Email        string    `json:"email" gorm:"type:varchar(100);not null;uniqueIndex"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Role         Role      `json:"role" gorm:"type:varchar(20);not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRegister struct {
	Name      string  `json:"name" binding:"required,min=2,max=100"`
	Email     string  `json:"email" binding:"required,email,max=255"`
	Password  string  `json:"password" binding:"required,min=8,max=255"`
	AdminCode *string `json:"admin_code"`
}

type UserLogin struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
