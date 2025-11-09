package service

// import (
// 	"context"
// 	"os"
// 	"path/filepath"
// 	"strconv"
// 	"testing"

// 	"github.com/rs/zerolog"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/require"

// 	"github.com/ummuys/reportify/internal/mocks"
// 	"github.com/ummuys/reportify/internal/models"
// )

// func newSvc(t *testing.T) (ReportService, *mocks.MockRepDB, *mocks.MockRepConv, *mocks.MockRepCache) {
// 	t.Helper()
// 	var zl zerolog.Logger
// 	db := &mocks.MockRepDB{}
// 	conv := &mocks.MockRepConv{}
// 	cache := &mocks.MockRepCache{}
// 	s := NewReportService(&zl, db, conv, cache)
// 	return s, db, conv, cache
// }

// func createTmpFile(t *testing.T) *os.File {
// 	t.Helper()
// 	dir := t.TempDir()
// 	fp := filepath.Join(dir, "out.text")
// 	f, err := os.Create(fp)
// 	require.NoError(t, err)
// 	t.Cleanup(func() { _ = f.Close() })
// 	return f
// }

// func TestCreateReport_Success_All(t *testing.T) {
// 	formats := []struct {
// 		name   string
// 		format string
// 		params models.ReportParams
// 		setup  func(conv *mocks.MockRepConv, headers []string, rows [][]any, f *os.File, params models.ReportParams)
// 	}{
// 		{
// 			name:   "convert to pdf",
// 			format: "pdf",
// 			params: models.ReportParams{},
// 			setup: func(conv *mocks.MockRepConv, headers []string, rows [][]any, f *os.File, _ models.ReportParams) {
// 				conv.On("ToPDF", headers, rows, f).Return(nil).Once()
// 			},
// 		},
// 		{
// 			name:   "convert to json",
// 			format: "json",
// 			params: models.ReportParams{},
// 			setup: func(conv *mocks.MockRepConv, headers []string, rows [][]any, f *os.File, _ models.ReportParams) {
// 				conv.On("ToJSON", headers, rows, f).Return(nil).Once()
// 			},
// 		},
// 		{
// 			name:   "convert to XLSX",
// 			format: "xlsx",
// 			params: models.ReportParams{},
// 			setup: func(conv *mocks.MockRepConv, headers []string, rows [][]any, f *os.File, _ models.ReportParams) {
// 				conv.On("ToXLSX", headers, rows, f).Return(nil).Once()
// 			},
// 		},
// 		{
// 			name:   "convert to CSV",
// 			format: "csv",
// 			params: models.ReportParams{CSVSep: ';'},
// 			setup: func(conv *mocks.MockRepConv, headers []string, rows [][]any, f *os.File, p models.ReportParams) {
// 				// важно: ToCSV принимает ещё sep
// 				conv.On("ToCSV", headers, rows, f, rune(p.CSVSep)).Return(nil).Once()
// 			},
// 		},
// 	}

// 	for _, tc := range formats {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ctx := context.Background()
// 			svc, mDB, mConv, mCch := newSvc(t)
// 			f := createTmpFile(t)

// 			headers := []string{"h"}
// 			rows := [][]any{{1}, {2}}
// 			sql := "hello!"
// 			userID := int64(52)

// 			// DB expectation
// 			mDB.On("CreateReport", mock.Anything, sql).Return(headers, rows, nil).Once()

// 			// converter expectation
// 			tc.setup(mConv, headers, rows, f, tc.params)

// 			// cache expectation (ключ — user_id как строка)
// 			mCch.On("Set", mock.Anything, strconv.FormatInt(userID, 10), sql).Return(nil).Once()

// 			params := tc.params
// 			params.Sql = sql

// 			err := svc.CreateReport(ctx, userID, params, f, tc.format)
// 			require.NoError(t, err)

// 			mDB.AssertExpectations(t)
// 			mConv.AssertExpectations(t)
// 			mCch.AssertExpectations(t)
// 		})
// 	}
// }

// func TestCreateReport_DBError(t *testing.T) {
// 	ctx := context.Background()
// 	svc, mDB, mConv, mCch := newSvc(t)
// 	f := createTmpFile(t)

// 	sql := "bad sql"
// 	userID := int64(1)

// 	mDB.On("CreateReport", mock.Anything, sql).Return(nil, nil, assertAnError()).Once()

// 	err := svc.CreateReport(ctx, userID, models.ReportParams{Sql: sql}, f, "pdf")
// 	require.Error(t, err)

// 	mConv.AssertNotCalled(t, "ToPDF", mock.Anything, mock.Anything, mock.Anything)
// 	mConv.AssertNotCalled(t, "ToXLSX", mock.Anything, mock.Anything, mock.Anything)
// 	mConv.AssertNotCalled(t, "ToJSON", mock.Anything, mock.Anything, mock.Anything)
// 	mConv.AssertNotCalled(t, "ToCSV", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

// 	mCch.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)

// 	mDB.AssertExpectations(t)
// }

// func TestCreateReport_ConverterError_XLSX(t *testing.T) {
// 	ctx := context.Background()
// 	svc, mDB, mConv, mCch := newSvc(t)
// 	f := createTmpFile(t)

// 	sql := "SELECT 1"
// 	userID := int64(7)
// 	headers := []string{"x"}
// 	rows := [][]any{{1}}

// 	mDB.On("CreateReport", mock.Anything, sql).Return(headers, rows, nil).Once()
// 	mConv.On("ToXLSX", headers, rows, f).Return(assertAnError()).Once()

// 	err := svc.CreateReport(ctx, userID, models.ReportParams{Sql: sql}, f, "xlsx")
// 	require.Error(t, err)

// 	mCch.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)

// 	mDB.AssertExpectations(t)
// 	mConv.AssertExpectations(t)
// }

// func TestCreateReport_CacheError_IsIgnored(t *testing.T) {
// 	ctx := context.Background()
// 	svc, mDB, mConv, mCch := newSvc(t)
// 	f := createTmpFile(t)

// 	sql := "SELECT 1"
// 	userID := int64(99)
// 	headers := []string{"x"}
// 	rows := [][]any{{1}}

// 	mDB.On("CreateReport", mock.Anything, sql).Return(headers, rows, nil).Once()
// 	mConv.On("ToJSON", headers, rows, f).Return(nil).Once()
// 	// Кэш возвращает ошибку, но метод должен вернуть nil
// 	mCch.On("Set", mock.Anything, strconv.FormatInt(userID, 10), sql).Return(assertAnError()).Once()

// 	err := svc.CreateReport(ctx, userID, models.ReportParams{Sql: sql}, f, "json")
// 	require.NoError(t, err)

// 	mDB.AssertExpectations(t)
// 	mConv.AssertExpectations(t)
// 	mCch.AssertExpectations(t)
// }

// func TestCreateReport_UnknownFormat_CurrentBehavior(t *testing.T) {
// 	// Текущее поведение: конвертер не вызывается, кэш вызывается, ошибки нет.
// 	ctx := context.Background()
// 	svc, mDB, mConv, mCch := newSvc(t)
// 	f := createTmpFile(t)

// 	sql := "SELECT 1"
// 	userID := int64(1)
// 	headers := []string{"x"}
// 	rows := [][]any{{1}}

// 	mDB.On("CreateReport", mock.Anything, sql).Return(headers, rows, nil).Once()
// 	// конвертер НЕ должен вызываться
// 	mCch.On("Set", mock.Anything, strconv.FormatInt(userID, 10), sql).Return(nil).Once()

// 	err := svc.CreateReport(ctx, userID, models.ReportParams{Sql: sql}, f, "unknown")
// 	require.NoError(t, err)

// 	mConv.AssertNotCalled(t, "ToPDF", mock.Anything, mock.Anything, mock.Anything)
// 	mConv.AssertNotCalled(t, "ToXLSX", mock.Anything, mock.Anything, mock.Anything)
// 	mConv.AssertNotCalled(t, "ToJSON", mock.Anything, mock.Anything, mock.Anything)
// 	mConv.AssertNotCalled(t, "ToCSV", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

// 	mDB.AssertExpectations(t)
// 	mCch.AssertExpectations(t)
// }

// func assertAnError() error {
// 	return &mockError{":)"}
// }

// type mockError struct{ s string }

// func (e *mockError) Error() string { return e.s }
