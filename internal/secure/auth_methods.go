package secure

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type tokMan struct {
}

func NewTokenManager() TokenManager {
	return &tokMan{}
}

func (tm tokMan) GenerateRefreshToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": username,
		"exp":     time.Now().Add(time.Hour * 144).Unix(), // Срок действия — 144 часа
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return refresh.SignedString([]byte(secret_refresh))
}

func (tm tokMan) GenerateAccessToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": username,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Срок действия — 24 часа
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return access.SignedString([]byte(secret_access))
}

func (tm tokMan) ValidateToken(rawToken string, mode bool) (jwt.MapClaims, error) {
	claims, err := getClaims(rawToken, mode)
	return claims, err
}
