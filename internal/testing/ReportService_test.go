package testing

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/ummuys/reportify/internal/mocks"
	"github.com/ummuys/reportify/internal/service"
)

func TestCreateReport_OK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := zerolog.Nop()
	dbMock := mocks.NewReportDB(t)
	cacheMock := mocks.NewReportCache(t)
	convMock := mocks.NewReportConvert(t)

	srv := service.NewReportService(&logger, dbMock, convMock, cacheMock)

	sql := "select 2"
	userID := int64(123)
	headers := []string{"x"}
	rows := [][]any{{2}}

	tmpDir := t.TempDir()
	outFile, err := os.Create(filepath.Join(tmpDir, "report.pdf"))
	require.NoError(t, err)
	defer outFile.Close()

	// Гарантируем порядок вызовов
	mock.InOrder(
		dbMock.On("ExecQuery", ctx, sql).
			Return(headers, rows, nil).
			Once(),
		convMock.On("ToPDF", headers, rows, outFile).
			Return(nil).
			Once(),
		cacheMock.On("SetQuery", ctx, "123", sql).
			Return(nil).
			Once(),
	)

	err = srv.CreateReport(ctx, userID, sql, outFile)
	require.NoError(t, err)

	dbMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
	cacheMock.AssertExpectations(t)
}

func TestCreateReport_NegativeCases(t *testing.T) {
	t.Parallel()

	type negCase struct {
		name          string
		sql           string
		setup         func(ctx context.Context, db *mocks.ReportDB, conv *mocks.ReportConvert, chc *mocks.ReportCache, f *os.File)
		expectNoConv  bool
		expectNoCache bool
	}

	cases := []negCase{
		{
			name: "SQLInjection: checkQuery падает до БД",
			sql:  "select 1",
			setup: func(ctx context.Context, db *mocks.ReportDB, conv *mocks.ReportConvert, chc *mocks.ReportCache, f *os.File) {
			},
			expectNoConv:  true,
			expectNoCache: true,
		},
		{
			name: "DBErr: ExecQuery возвращает ошибку",
			sql:  "select 2",
			setup: func(ctx context.Context, db *mocks.ReportDB, conv *mocks.ReportConvert, chc *mocks.ReportCache, f *os.File) {
				db.On("ExecQuery", ctx, "select 2").
					Return([]string{}, [][]any{}, errors.New("db err")).
					Once()
			},
			expectNoConv:  true,
			expectNoCache: true,
		},
		{
			name: "ToPDFErr: конвертер падает — кэш не вызываем",
			sql:  "select 2",
			setup: func(ctx context.Context, db *mocks.ReportDB, conv *mocks.ReportConvert, chc *mocks.ReportCache, f *os.File) {
				headers := []string{"c"}
				rows := [][]any{{2}}

				mock.InOrder(
					db.On("ExecQuery", ctx, "select 2").
						Return(headers, rows, nil).
						Once(),
					conv.On("ToPDF", headers, rows, f).
						Return(errors.New("pdf conv failed")).
						Once(),
				)
			},
			expectNoConv:  false,
			expectNoCache: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			logger := zerolog.Nop()
			dbMock := mocks.NewReportDB(t)
			cacheMock := mocks.NewReportCache(t)
			convMock := mocks.NewReportConvert(t)

			srv := service.NewReportService(&logger, dbMock, convMock, cacheMock)

			tmpDir := t.TempDir()
			outFile, err := os.Create(filepath.Join(tmpDir, "report.pdf"))
			require.NoError(t, err)
			defer outFile.Close()

			tc.setup(ctx, dbMock, convMock, cacheMock, outFile)

			err = srv.CreateReport(ctx, 123, tc.sql, outFile)
			require.Error(t, err)

			if tc.expectNoConv {
				convMock.AssertNotCalled(t, "ToPDF", mock.Anything, mock.Anything, mock.Anything)
			}
			if tc.expectNoCache {
				cacheMock.AssertNotCalled(t, "SetQuery", mock.Anything, mock.Anything, mock.Anything)
			}

			dbMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
			cacheMock.AssertExpectations(t)
		})
	}
}
