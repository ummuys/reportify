package models

type CreateReport struct {
	Sql string `json:"sql"`
}

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
