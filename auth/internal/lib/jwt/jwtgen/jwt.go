package jwtgen

import (
	"auth/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewToken(user *models.User, secret string, duration time.Duration) (string, error) {
	claims := make(map[string]any)

	claims["uid"] = user.ID
	claims["exp"] = time.Now().Add(duration).Unix()

	return generateToken(claims, secret)
}

func NewRefreshToken(user *models.User, secret string, duration time.Duration) (string, error) {
	claims := make(map[string]any)

	claims["uid"] = user.ID
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["type"] = "refresh"

	return generateToken(claims, secret)
}

func generateToken(claims map[string]any, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
