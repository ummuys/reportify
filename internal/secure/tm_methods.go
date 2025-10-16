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

func (tm *tokMan) GenerateRefreshToken(user_id int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user_id,
		"exp":     time.Now().Add(time.Hour * 144).Unix(), // Срок действия — 144 часа
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return refresh.SignedString([]byte(secret_refresh))
}

func (tm *tokMan) GenerateAccessToken(user_id int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user_id,
		"exp":     time.Now().Add(time.Hour * 2).Unix(), // Срок действия — 2 часа
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return access.SignedString([]byte(secret_access))
}

func (tm *tokMan) ValidateToken(rawToken string, mode bool) (jwt.MapClaims, error) {
	claims, err := getClaims(rawToken, mode)
	return claims, err
}
