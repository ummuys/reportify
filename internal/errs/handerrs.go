package errs

import "errors"

var (

	// SERVER ERR
	ErrInternalServer  = errors.New("something wrong with server, try again later")
	ErrBadRefreshToken = errors.New("bad refresh token")
	ErrBadAccessToken  = errors.New("bad access token")
	ErrInvalidJSON     = errors.New("invalid JSON format")

	// ANOTHER
	ErrUserUnauthorized   = errors.New("you need to auth")
	ErrInvalidCredentials = errors.New("user or password not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmptyUsername      = errors.New("empty username")
	ErrEmptySchemaName    = errors.New("schema name requeired")
	ErrEmptyTableName     = errors.New("table name required")
)
