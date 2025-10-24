package secure

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/ummuys/reportify/internal/config"
)

var cfg config.TMConfig

type tokMan struct {
}

func NewTokenManager() (TokenManager, error) {
	c, err := config.ParseTMConfig()
	if err != nil {
		return nil, err
	}
	cfg = c
	return &tokMan{}, nil
}

func (tm *tokMan) GenerateRefreshToken(user_id int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user_id,
		"exp":     cfg.RefreshTokenLimit, // Срок действия — 144 часа
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return refresh.SignedString([]byte(cfg.RefreshSecret))
}

func (tm *tokMan) GenerateAccessToken(user_id int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user_id,
		"exp":     cfg.AccessTokenLimit, // Срок действия — 2 часа
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return access.SignedString([]byte(cfg.AccessSecret))
}

func (tm *tokMan) ValidateToken(rawToken string, mode bool) (jwt.MapClaims, error) {
	claims, err := getClaims(rawToken, mode)
	return claims, err
}
