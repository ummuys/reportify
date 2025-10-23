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
	var (
		out1 []string
		out2 [][]any
		ok   bool
	)
	if out1, ok = args.Get(0).([]string); !ok {
		out1 = nil
	}
	if out2, ok = args.Get(1).([][]any); !ok {
		out2 = nil
	}

	return out1, out2, args.Error(2)
}
