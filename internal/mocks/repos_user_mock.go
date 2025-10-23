package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockUDB struct {
	mock.Mock
}

func (m *MockUDB) CreateUser(pCtx context.Context, username string, hashPassword string) error {
	args := m.Called(pCtx, username, hashPassword)
	return args.Error(0)
}
func (m *MockUDB) GetPassword(pCtx context.Context, username string) (int64, string, error) {
	args := m.Called(pCtx, username)
	return args.Get(0).(int64), args.String(1), args.Error(2)
}
func (m *MockUDB) Exists(pCtx context.Context, username string) error {
	args := m.Called(pCtx, username)
	return args.Error(0)
}
