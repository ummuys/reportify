package webdto

type GetUsersResponse struct {
	Users []UserResponse `json:"users"`
}

type UserResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
