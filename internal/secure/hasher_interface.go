package secure

type PasswordHasher interface {
	Hash(password string) (string, error)
	CheckHash(password, hash string) bool
}
