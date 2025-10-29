package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRepCache struct {
	mock.Mock
}

func (m *MockRepCache) Init(pCtx context.Context, queries map[string][]string) error {
	args := m.Called(pCtx, queries)
	return args.Error(0)
}

func (m *MockRepCache) Set(pCtx context.Context, key string, value string) error {
	args := m.Called(pCtx, key, value)
	return args.Error(0)
}

func (m *MockRepCache) Get(pCtx context.Context, key string) ([]string, error) {
	args := m.Called(pCtx, key)
	var out []string
	if v := args.Get(0); v != nil {
		out = v.([]string)
	}
	return out, args.Error(1)
}

func (m *MockRepCache) GetAll(pCtx context.Context) (map[string][]string, error) {
	args := m.Called(pCtx)
	var out map[string][]string
	if v := args.Get(0); v != nil {
		out = v.(map[string][]string)
	}
	return out, args.Error(1)
}

func (m *MockRepCache) Delete(pCtx context.Context, key string, value string) error {
	args := m.Called(pCtx, key, value)
	return args.Error(0)
}

func (m *MockRepCache) DeleteAll(pCtx context.Context, key string) error {
	args := m.Called(pCtx, key)
	return args.Error(0)
}
