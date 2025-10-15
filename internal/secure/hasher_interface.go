package secure

type PasswordHasher interface {
	Hash(password string) (string, error)
	ChechHash(password, hash string) bool
}
