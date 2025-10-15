package web

// BASE PATH

// REPORT
const (
	CreateReportPath = "report/:format"
)

// Metadata
const (
	GetSchemasPath     = "db/schemas"
	GetTablesPath      = "db/tables"
	GetColumnsPath     = "db/columns"
	GetCacheQuerysPath = "cache"
	ClearHistoryPath   = "cache/clear"
)

// AUTH
const (
	AuthPath           = "secure/auth"
	GetAccessTokenPath = "secure/access"
)
