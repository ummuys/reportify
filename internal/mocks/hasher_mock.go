package mocks

import "github.com/stretchr/testify/mock"

type MockHasher struct {
	mock.Mock
}

func (m *MockHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockHasher) CheckHash(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}
