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

// AUTH
const (
	AuthPath           = "secure/auth"
	GetAccessTokenPath = "secure/access"
)
