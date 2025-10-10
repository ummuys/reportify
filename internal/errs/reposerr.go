package errs

import "errors"

var (
	ErrBadData     = errors.New("incorrect login or password")
	ErrUsernameUnq = errors.New("username alredy is taken")
)
