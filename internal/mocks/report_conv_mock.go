package mocks

import (
	"os"

	"github.com/stretchr/testify/mock"
)

type MockRepConv struct {
	mock.Mock
}

func (m *MockRepConv) ToDOCX(headers []string, data [][]any, f *os.File) error {
	args := m.Called(headers, data, f)
	return args.Error(0)
}

func (m *MockRepConv) ToPDF(headers []string, data [][]any, f *os.File) error {
	args := m.Called(headers, data, f)
	return args.Error(0)
}

func (m *MockRepConv) ToXLSX(headers []string, data [][]any, f *os.File) error {
	args := m.Called(headers, data, f)
	return args.Error(0)
}

func (m *MockRepConv) ToJSON(headers []string, data [][]any, f *os.File) error {
	args := m.Called(headers, data, f)
	return args.Error(0)
}

func (m *MockRepConv) ToCSV(headers []string, data [][]any, f *os.File, sep rune) error {
	args := m.Called(headers, data, f, sep)
	return args.Error(0)
}
