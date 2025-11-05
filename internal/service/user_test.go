package service

// import (
// 	"context"
// 	"testing"

// 	"github.com/rs/zerolog"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/require"

// 	"github.com/ummuys/reportify/internal/errs"
// 	"github.com/ummuys/reportify/internal/mocks"
// )

// // ---- helpers ----

// func newUserSvc(t *testing.T) (*uSrv, *mocks.MockUDB, *mocks.MockHasher) {
// 	t.Helper()
// 	var zl zerolog.Logger
// 	db := &mocks.MockUDB{}
// 	ph := &mocks.MockHasher{}
// 	s := NewUserService(&zl, db, ph).(*uSrv)
// 	return s, db, ph
// }

// // ---- Create ----

// func TestUser_Create_Success(t *testing.T) {
// 	ctx := context.Background()
// 	s, db, ph := newUserSvc(t)

// 	username := "vasya"
// 	raw := "pass123"
// 	hashed := "h::pass123"

// 	// Exists OK
// 	db.On("Exists", mock.Anything, username).Return(nil).Once()
// 	// Hash OK
// 	ph.On("Hash", raw).Return(hashed, nil).Once()
// 	// CreateUser OK
// 	db.On("CreateUser", mock.Anything, username, hashed).Return(nil).Once()

// 	err := s.CreateUser(ctx, username, raw, "user")
// 	require.NoError(t, err)

// 	db.AssertExpectations(t)
// 	ph.AssertExpectations(t)
// }

// func TestUser_Create_ExistsError(t *testing.T) {
// 	ctx := context.Background()
// 	s, db, ph := newUserSvc(t)

// 	username := "vasya"
// 	raw := "pass123"
// 	someErr := anyErr("exists failed")

// 	db.On("Exists", mock.Anything, username).Return(someErr).Once()

// 	err := s.Create(ctx, username, raw, "user")
// 	require.ErrorIs(t, err, someErr)

// 	// Hash/CreateUser не должны вызываться
// 	ph.AssertNotCalled(t, "Hash", mock.Anything)
// 	db.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything, mock.Anything)

// 	db.AssertExpectations(t)
// }

// func TestUser_Create_HashError(t *testing.T) {
// 	ctx := context.Background()
// 	s, db, ph := newUserSvc(t)

// 	username := "vasya"
// 	raw := "pass123"
// 	someErr := anyErr("hash failed")

// 	db.On("Exists", mock.Anything, username).Return(nil).Once()
// 	ph.On("Hash", raw).Return("", someErr).Once()

// 	err := s.Create(ctx, username, raw, "user")
// 	require.ErrorIs(t, err, someErr)

// 	db.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything, mock.Anything)

// 	db.AssertExpectations(t)
// 	ph.AssertExpectations(t)
// }

// func TestUser_Create_CreateUserError(t *testing.T) {
// 	ctx := context.Background()
// 	s, db, ph := newUserSvc(t)

// 	username := "vasya"
// 	raw := "pass123"
// 	hashed := "h::pass123"
// 	someErr := anyErr("insert failed")

// 	db.On("Exists", mock.Anything, username).Return(nil).Once()
// 	ph.On("Hash", raw).Return(hashed, nil).Once()
// 	db.On("CreateUser", mock.Anything, username, hashed).Return(someErr).Once()

// 	err := s.Create(ctx, username, raw, "user")
// 	require.ErrorIs(t, err, someErr)

// 	db.AssertExpectations(t)
// 	ph.AssertExpectations(t)
// }

// // ---- CheckCredentials ----

// func TestUser_CheckCredentials_Success(t *testing.T) {
// 	ctx := context.Background()
// 	s, db, ph := newUserSvc(t)

// 	username := "vasya"
// 	raw := "pass123"
// 	storedHash := "h::pass123"
// 	uid := int64(42)

// 	db.On("Exists", mock.Anything, username).Return(nil).Once()
// 	db.On("GetPassword", mock.Anything, username).Return(uid, storedHash, nil).Once()
// 	ph.On("CheckHash", raw, storedHash).Return(true).Once()

// 	gotUID, _, err := s.CheckCredentials(ctx, username, raw)
// 	require.NoError(t, err)
// 	require.Equal(t, uid, gotUID)

// 	db.AssertExpectations(t)
// 	ph.AssertExpectations(t)
// }

// func TestUser_CheckCredentials_ExistsError(t *testing.T) {
// 	ctx := context.Background()
// 	s, db, ph := newUserSvc(t)

// 	username := "vasya"
// 	raw := "pass123"
// 	someErr := anyErr("exists failed")

// 	db.On("Exists", mock.Anything, username).Return(someErr).Once()

// 	uid, _, err := s.CheckCredentials(ctx, username, raw)
// 	require.Zero(t, uid)
// 	require.ErrorIs(t, err, someErr)

// 	db.AssertNotCalled(t, "GetPassword", mock.Anything, mock.Anything)
// 	ph.AssertNotCalled(t, "CheckHash", mock.Anything, mock.Anything)

// 	db.AssertExpectations(t)
// }

// func TestUser_CheckCredentials_GetPasswordError(t *testing.T) {
// 	ctx := context.Background()
// 	s, db, ph := newUserSvc(t)

// 	username := "vasya"
// 	raw := "pass123"
// 	someErr := anyErr("select failed")

// 	db.On("Exists", mock.Anything, username).Return(nil).Once()
// 	db.On("GetPassword", mock.Anything, username).Return(int64(0), "", someErr).Once()

// 	uid, _, err := s.CheckCredentials(ctx, username, raw)
// 	require.Zero(t, uid)
// 	require.ErrorIs(t, err, someErr)

// 	ph.AssertNotCalled(t, "CheckHash", mock.Anything, mock.Anything)

// 	db.AssertExpectations(t)
// }

// func TestUser_CheckCredentials_InvalidPassword(t *testing.T) {
// 	ctx := context.Background()
// 	s, db, ph := newUserSvc(t)

// 	username := "vasya"
// 	raw := "badpass"
// 	storedHash := "h::good"
// 	uid := int64(42)

// 	db.On("Exists", mock.Anything, username).Return(nil).Once()
// 	db.On("GetPassword", mock.Anything, username).Return(uid, storedHash, nil).Once()
// 	ph.On("CheckHash", raw, storedHash).Return(false).Once()

// 	gotUID, _, err := s.CheckCredentials(ctx, username, raw)
// 	require.Zero(t, gotUID)
// 	require.ErrorIs(t, err, errs.ErrInvalidCredentials)

// 	db.AssertExpectations(t)
// 	ph.AssertExpectations(t)
// }
