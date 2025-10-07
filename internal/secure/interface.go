package secure

import "github.com/golang-jwt/jwt/v5"

type TokenManager interface {
	GenerateRefreshToken(username string) (string, error)
	GenerateAccessToken(username string) (string, error)
	ValidateToken(rawToken string, mode bool) (jwt.MapClaims, error) // mode == true - access, mode == false - refresh
}

var secret_access = "zopa"
var secret_refresh = "popa"
