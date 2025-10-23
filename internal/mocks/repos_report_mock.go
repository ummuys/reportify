package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRepDB struct {
	mock.Mock
}

func (m *MockRepDB) CreateReport(pCtx context.Context, script string) ([]string, [][]any, error) {
	args := m.Called(pCtx, script)
	return args.Get(0).([]string), args.Get(1).([][]any), args.Error(2)
}
