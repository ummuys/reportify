package secure

import (
	"golang.org/x/crypto/bcrypt"
)

type bcryptH struct{}

func NewPasswordHasher() PasswordHasher {
	return &bcryptH{}
}

func (b *bcryptH) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (b *bcryptH) CheckHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
