package service

// import (
// 	"context"
// 	"testing"

// 	"github.com/rs/zerolog"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/require"

// 	"github.com/ummuys/reportify/internal/mocks"
// 	"github.com/ummuys/reportify/internal/models"
// )

// // --- helpers ---

// func newMDService(t *testing.T) (*mdService, *mocks.MockMDDB, *mocks.MockRepCache) {
// 	t.Helper()
// 	var zl zerolog.Logger
// 	db := &mocks.MockMDDB{}
// 	cache := &mocks.MockRepCache{}
// 	s := NewMetadataService(&zl, db, cache).(*mdService)
// 	return s, db, cache
// }

// // для проверки без учёта порядка (map iteration в Go не детерминированна)
// func schemasToMap(ls *models.ListSchemas) map[string]string {
// 	m := make(map[string]string, len(ls.Schemas))
// 	for _, s := range ls.Schemas {
// 		m[s.Name] = s.Comment
// 	}
// 	return m
// }
// func tablesToMap(lt *models.ListTables) map[string]string {
// 	m := make(map[string]string, len(lt.Tables))
// 	for _, s := range lt.Tables {
// 		m[s.Name] = s.Comment
// 	}
// 	return m
// }
// func columnsToMap(lc *models.ListColumns) map[string]string {
// 	m := make(map[string]string, len(lc.Columns))
// 	for _, s := range lc.Columns {
// 		m[s.Name] = s.Comment
// 	}
// 	return m
// }

// // --- GetSchemas ---

// func TestMetadata_GetSchemas_Success(t *testing.T) {
// 	ctx := context.Background()
// 	svc, mdb, _ := newMDService(t)

// 	in := map[string]string{
// 		"public":  "default schema",
// 		"account": "users & billing",
// 	}
// 	mdb.On("GetSchemas", mock.Anything).Return(in, nil).Once()

// 	got, err := svc.GetSchemas(ctx)
// 	require.NoError(t, err)
// 	require.NotNil(t, got)
// 	require.Equal(t, len(in), len(got.Schemas))

// 	gotMap := schemasToMap(got)
// 	require.Equal(t, in, gotMap)

// 	mdb.AssertExpectations(t)
// }

// func TestMetadata_GetSchemas_DBError(t *testing.T) {
// 	ctx := context.Background()
// 	svc, mdb, _ := newMDService(t)

// 	mdb.On("GetSchemas", mock.Anything).Return(nil, anyErr("db fail")).Once()

// 	got, err := svc.GetSchemas(ctx)
// 	require.Nil(t, got)
// 	require.Error(t, err)

// 	mdb.AssertExpectations(t)
// }

// // --- GetTables ---

// func TestMetadata_GetTables_Success(t *testing.T) {
// 	ctx := context.Background()
// 	svc, mdb, _ := newMDService(t)

// 	schema := "public"
// 	in := map[string]string{
// 		"users":    "users table",
// 		"invoices": "invoices table",
// 	}
// 	mdb.On("GetTables", mock.Anything, schema).Return(in, nil).Once()

// 	got, err := svc.GetTables(ctx, schema)
// 	require.NoError(t, err)
// 	require.NotNil(t, got)
// 	require.Equal(t, len(in), len(got.Tables))

// 	gotMap := tablesToMap(got)
// 	require.Equal(t, in, gotMap)

// 	mdb.AssertExpectations(t)
// }

// func TestMetadata_GetTables_DBError(t *testing.T) {
// 	ctx := context.Background()
// 	svc, mdb, _ := newMDService(t)

// 	schema := "public"
// 	mdb.On("GetTables", mock.Anything, schema).Return(nil, anyErr("db fail")).Once()

// 	got, err := svc.GetTables(ctx, schema)
// 	require.Nil(t, got)
// 	require.Error(t, err)

// 	mdb.AssertExpectations(t)
// }

// // --- GetColumns ---

// func TestMetadata_GetColumns_Success(t *testing.T) {
// 	ctx := context.Background()
// 	svc, mdb, _ := newMDService(t)

// 	schema := "public"
// 	table := "users"
// 	in := map[string]string{
// 		"id":        "primary key",
// 		"email":     "user email",
// 		"createdAt": "creation time",
// 	}
// 	mdb.On("GetColumns", mock.Anything, schema, table).Return(in, nil).Once()

// 	got, err := svc.GetColumns(ctx, schema, table)
// 	require.NoError(t, err)
// 	require.NotNil(t, got)
// 	require.Equal(t, len(in), len(got.Columns))

// 	gotMap := columnsToMap(got)
// 	require.Equal(t, in, gotMap)

// 	mdb.AssertExpectations(t)
// }

// func TestMetadata_GetColumns_DBError(t *testing.T) {
// 	ctx := context.Background()
// 	svc, mdb, _ := newMDService(t)

// 	schema := "public"
// 	table := "users"
// 	mdb.On("GetColumns", mock.Anything, schema, table).Return(nil, anyErr("db fail")).Once()

// 	got, err := svc.GetColumns(ctx, schema, table)
// 	require.Nil(t, got)
// 	require.Error(t, err)

// 	mdb.AssertExpectations(t)
// }

// // --- GetQueries (cache.Get) ---

// func TestMetadata_GetQueries_Success(t *testing.T) {
// 	ctx := context.Background()
// 	svc, _, cache := newMDService(t)

// 	key := "42"
// 	val := []string{"select 1", "select 2"}
// 	cache.On("Get", mock.Anything, key).Return(val, nil).Once()

// 	got, err := svc.GetQueries(ctx, key)
// 	require.NoError(t, err)
// 	require.Equal(t, val, got)

// 	cache.AssertExpectations(t)
// }

// func TestMetadata_GetQueries_Error(t *testing.T) {
// 	ctx := context.Background()
// 	svc, _, cache := newMDService(t)

// 	key := "42"
// 	cache.On("Get", mock.Anything, key).Return(nil, anyErr("cache fail")).Once()

// 	got, err := svc.GetQueries(ctx, key)
// 	require.Nil(t, got)
// 	require.Error(t, err)

// 	cache.AssertExpectations(t)
// }

// // --- DeleteAllQueries (cache.DeleteAll) ---

// func TestMetadata_DeleteAllQueries_Success(t *testing.T) {
// 	ctx := context.Background()
// 	svc, _, cache := newMDService(t)

// 	key := "42"
// 	cache.On("DeleteAll", mock.Anything, key).Return(nil).Once()

// 	err := svc.DeleteAllQueries(ctx, key)
// 	require.NoError(t, err)

// 	cache.AssertExpectations(t)
// }

// func TestMetadata_DeleteAllQueries_Error(t *testing.T) {
// 	ctx := context.Background()
// 	svc, _, cache := newMDService(t)

// 	key := "42"
// 	cache.On("DeleteAll", mock.Anything, key).Return(anyErr("cache fail")).Once()

// 	err := svc.DeleteAllQueries(ctx, key)
// 	require.Error(t, err)

// 	cache.AssertExpectations(t)
// }
