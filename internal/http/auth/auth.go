package auth

import (
	"avito-tech-winter-2025/internal/config"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Manager struct {
	secretKey string
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func NewManager(cfg *config.Config) *Manager {
	if cfg == nil {
		panic("nil config provided to auth manager")
	}
	return &Manager{secretKey: cfg.Jwt.SecretKey}
}

func (m *Manager) NewJwt(userID int, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
