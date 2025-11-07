package repository

import "context"

// TO FIX: MAKE BETTER ERR CHECKER
// TO FIX: MAKE BETTER ERR CHECKER
// TO FIX: MAKE BETTER ERR CHECKER
// TO FIX: MAKE BETTER ERR CHECKER
type UserDB interface {
	GetUsers(pCtx context.Context) ([][]string, error)
	CreateUser(pCtx context.Context, username string, hashPassword string, role string) error
	DeleteUser(pCtx context.Context, username string) error
	CheckCredentials(pCtx context.Context, username string) (int64, string, string, error)
	CheckRole(pCtx context.Context, role string) error
	CheckUser(pCtx context.Context, username string) error
}
