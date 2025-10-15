package errs

import "errors"

var (
	ErrInvalidCredentials   = errors.New("incorrect login or password")
	ErrUsernameAlredyExists = errors.New("username alredy is taken")
)
