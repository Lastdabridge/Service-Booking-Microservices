package auth

import (
	"errors"
	"os"
	"time"

	"github.com/Lastdabridge/Service-Booking-Microservices/services/auth-service/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint       `json:"user_id"`
	Role   model.Role `json:"role"`
	jwt.RegisteredClaims
}

func getSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}

	return []byte(secret), nil
}

func GenerateToken(user *model.User) (string, error) {
	secret, err := getSecret()
	if err != nil {
		return "", err
	}

	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(24 * time.Hour),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(secret))
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
