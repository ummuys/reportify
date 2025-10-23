package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockMDDB struct{ mock.Mock }

func (m *MockMDDB) GetSchemas(ctx context.Context) (map[string]string, error) {
	args := m.Called(ctx)
	var out map[string]string
	if v := args.Get(0); v != nil {
		out = v.(map[string]string)
	}
	return out, args.Error(1)
}

func (m *MockMDDB) GetTables(ctx context.Context, schemaName string) (map[string]string, error) {
	args := m.Called(ctx, schemaName)
	var out map[string]string
	if v := args.Get(0); v != nil {
		out = v.(map[string]string)
	}
	return out, args.Error(1)
}

func (m *MockMDDB) GetColumns(ctx context.Context, schemaName, tableName string) (map[string]string, error) {
	args := m.Called(ctx, schemaName, tableName)
	var out map[string]string
	if v := args.Get(0); v != nil {
		out = v.(map[string]string)
	}
	return out, args.Error(1)
}

func (m *MockMDDB) SetCacheQueries(ctx context.Context, cache map[string][]string) error {
	args := m.Called(ctx, cache)
	return args.Error(0)
}

func (m *MockMDDB) GetCacheQueries(ctx context.Context) (map[string][]string, error) {
	args := m.Called(ctx)
	var out map[string][]string
	if v := args.Get(0); v != nil {
		out = v.(map[string][]string)
	}
	return out, args.Error(1)
}
