package jwtTypes

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// Структура пользователя из БД
type User struct {
	ID       int    `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Password string `db:"password" json:"-"`
}

// Для логина
type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func CreateJWTToken(userID int, userName, jwtSecret string) (string, time.Time, error) {
	// Время истечения - через 24 часа
	expirationTime := time.Now().Add(24 * time.Hour)

	// Создаем claims
	claims := &jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "myapp",
	}

	// Создаем токен с claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}
