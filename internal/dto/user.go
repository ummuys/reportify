package dto

type GetUser struct {
	UserID   int64
	Username string
	Role     string
}

type UserCredentials struct {
	UserID   int64
	Password string
	Role     string
}

type CreateUser struct {
	Username string
	Password string
	Role     string
}

type UpdateUser struct {
	UserID   int64
	Username string
	Password string
	Role     string
}

type DeleteUser struct {
	Username string
}
