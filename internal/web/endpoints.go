package web

// BASE PATH

// REPORT
const (
	CreateReportPath = "report/:format"
)

// Metadata
const (
	GetSchemasPath       = "db/schemas"
	GetTablesPath        = "db/tables"
	GetColumnsPath       = "db/columns"
	GetAllQueriesPath    = "cache"
	DeleteAllQueriesPath = "cache/all"
	DeleteQueryPath      = "cache"
)

// ADMIN
const (

	// Collection
	GetUsersPath   = "/admin/users" // GET
	CreateUserPath = "/admin/users" // POST
	UpdateUserPath = "/admin/users" // PATCH

	// Item
	GetUserInfoPath = "/admin/users/:username" // GET
	DeleteUserPath  = "/admin/users/:username" // DELETE
)

// AUTH
const (
	AuthPath           = "secure/auth"
	GetAccessTokenPath = "secure/access"
)
