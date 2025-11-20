package secure

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ummuys/reportify/internal/config"
)

var cfg config.TMConfig

type tokMan struct{}

func NewTokenManager() (TokenManager, error) {
	c, err := config.ParseTMConfig()
	if err != nil {
		return nil, err
	}
	cfg = c
	return &tokMan{}, nil
}

func (tm *tokMan) GenerateRefreshToken(user_id int64, role string) (string, error) {
	claims := RefreshClaims{
		UserID: user_id,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.RefreshTokenLimit)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "app-token-manager",
			Audience:  []string{"app-users"},
		},
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return refresh.SignedString([]byte(cfg.RefreshSecret))
}

func (tm *tokMan) GenerateAccessToken(user_id int64, role string) (string, error) {
	claims := AccessClaims{
		UserID: user_id,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenLimit)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "app-token-manager",
			Audience:  []string{"app-users"},
		},
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return access.SignedString([]byte(cfg.AccessSecret))
}

func (tm *tokMan) ValidateAccessToken(rawToken string) (AccessClaims, error) {
	token, claims, err := unhashAccessToken(rawToken)
	if err != nil {
		return AccessClaims{}, err
	}
	if !token.Valid {
		return AccessClaims{}, fmt.Errorf("invalid token")
	}
	return claims, err
}

func (tm *tokMan) ValidateRefreshToken(rawToken string) (RefreshClaims, error) {
	token, claims, err := unhashRefreshToken(rawToken)
	if err != nil {
		return RefreshClaims{}, err
	}
	if !token.Valid {
		return RefreshClaims{}, fmt.Errorf("invalid token")
	}
	return claims, err
}

func unhashAccessToken(token string) (*jwt.Token, AccessClaims, error) {
	var claims AccessClaims
	jwtToken, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(cfg.AccessSecret), nil
	})
	return jwtToken, claims, err
}

func unhashRefreshToken(token string) (*jwt.Token, RefreshClaims, error) {
	var claims RefreshClaims
	jwtToken, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(cfg.RefreshSecret), nil
	})
	return jwtToken, claims, err
}
