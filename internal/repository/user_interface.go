package repository

import "context"

// TO FIX: MAKE BETTER ERR CHECKER
// TO FIX: MAKE BETTER ERR CHECKER
// TO FIX: MAKE BETTER ERR CHECKER
// TO FIX: MAKE BETTER ERR CHECKER
type UserDB interface {
	GetUsers(pCtx context.Context) ([][]any, error)
	CreateUser(pCtx context.Context, username string, hashPassword string, role string) (err error)
	UpdateUser(pCtx context.Context, userID int64, username string, hashPassword string, role string) (err error)
	DeleteUser(pCtx context.Context, username string) error
	CheckCredentials(pCtx context.Context, username string) (int64, string, string, error)
	ValidateRole(pCtx context.Context, role string) error
}
