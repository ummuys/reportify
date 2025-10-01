package repository

// COLUMNS QUERY
const (
	qColumnsWithComment = `
	SELECT
		a.attname                   AS column_name,
		COALESCE(d.description, '') AS comment
	FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_attribute  a ON a.attrelid = c.oid
	LEFT JOIN pg_description d
		ON d.objoid = c.oid AND d.objsubid = a.attnum
	WHERE n.nspname = $1         -- schema
		AND c.relname = $2          -- table
		AND a.attnum > 0
		AND NOT a.attisdropped;
	`
)

// TABLE QUERY
const (
	qTablesWithComment = `
	SELECT
		c.relname                                        AS table_name,
		COALESCE(obj_description(c.oid, 'pg_class'), '') AS comment
	FROM pg_class c
	JOIN pg_namespace n ON n.oid = c.relnamespace
	WHERE n.nspname = $1;
	`
)

// SCHEMA QUERY
const (
	qSchemaWithComment = `
	SELECT
		n.nspname                                         AS schema_name,
		COALESCE(obj_description(n.oid, 'pg_namespace'), '') AS comment
	FROM pg_namespace n
	WHERE n.nspname <> 'information_schema'
	AND n.nspname NOT LIKE 'pg\_%' ESCAPE '\';
	`
)
