package secure

type TokenManager interface {
	GenerateRefreshToken(user_id int64, role string) (string, error)
	GenerateAccessToken(user_id int64, role string) (string, error)
	ValidateAccessToken(rawToken string) (AccessClaims, error)
	ValidateRefreshToken(rawToken string) (RefreshClaims, error)
}
