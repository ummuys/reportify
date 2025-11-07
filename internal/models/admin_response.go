package models

type GetUsersResponse struct {
	List []UserResponse `json:"users"`
}

type UserResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}
