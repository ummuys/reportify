package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockUDB struct {
	mock.Mock
}

func (m *MockUDB) CreateUser(pCtx context.Context, username string, hashPassword, role string) error {
	args := m.Called(pCtx, username, hashPassword)
	return args.Error(0)
}

func (m *MockUDB) CheckCredentials(pCtx context.Context, username string) (int64, string, string, error) {
	args := m.Called(pCtx, username)
	return args.Get(0).(int64), args.String(1), args.String(2), args.Error(3)
}

func (m *MockUDB) Exists(pCtx context.Context, username string) error {
	args := m.Called(pCtx, username)
	return args.Error(0)
}
