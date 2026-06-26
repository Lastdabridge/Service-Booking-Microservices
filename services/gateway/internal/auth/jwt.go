package auth

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type Role string

const (
	RoleClient Role = "client"
	RoleAdmin  Role = "admin"
)

type Claims struct {
	UserID uint `json:"user_id"`
	Role   Role `json:"role"`
	jwt.RegisteredClaims
}

func ValidateToken(tokenString string) (*Claims, error) {
	secret, err := getSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func getSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}

	return []byte(secret), nil
}
