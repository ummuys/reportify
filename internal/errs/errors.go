package errs

import "errors"

// DB ERR
var (
	ErrInvalidCredentials   = errors.New("incorrect login or password")
	ErrUsernameAlredyExists = errors.New("username alredy is taken")
)
