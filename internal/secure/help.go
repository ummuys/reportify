package secure

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func unhashAccessToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret_access, nil
	})
}

func unhashRefreshToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret_refresh, nil
	})
}

func getClaims(rawToken string, mode bool) (jwt.MapClaims, error) {

	var (
		token *jwt.Token
		err   error
	)

	if mode {
		token, err = unhashAccessToken(rawToken)
		if err != nil {
			return nil, err
		}
	} else {
		token, err = unhashRefreshToken(rawToken)
		if err != nil {
			return nil, err
		}
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims")
	}

	return claims, nil

}
