package secure

import "github.com/golang-jwt/jwt/v5"

type TokenManager interface {
	GenerateRefreshToken(user_id int64) (string, error)
	GenerateAccessToken(user_id int64) (string, error)
	ValidateToken(rawToken string, mode bool) (jwt.MapClaims, error) // mode == true - access, mode == false - refresh
}

var secret_access = "zopa"
var secret_refresh = "popa"
