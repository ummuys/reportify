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
	AdminUsersBasePath = "/admin/users"

	// Collection
	GetAllUsersPath = "/admin/users" // GET
	CreateUserPath  = "/admin/users" // POST

	// Item
	GetUserInfoPath    = "/admin/users/:user_id" // GET
	ChangeUserInfoPath = "/admin/users/:user_id" // PATCH (или PUT)
	DeleteUserPath     = "/admin/users/:user_id" // DELETE
)

// AUTH
const (
	AuthPath           = "secure/auth"
	GetAccessTokenPath = "secure/access"
)
