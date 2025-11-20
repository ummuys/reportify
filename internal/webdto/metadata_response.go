package webdto

type ListTables struct {
	Tables []Table `json:"tables"`
}

type Table struct {
	Name    string `json:"table_name"`
	Comment string `json:"table_comm"`
}

type ListSchemas struct {
	Schemas []Schema `json:"schemas"`
}

type Schema struct {
	Name    string `json:"schema_name"`
	Comment string `json:"schema_comm"`
}

type ListColumns struct {
	Columns []Column `json:"columns"`
}

type Column struct {
	Name    string `json:"column_name"`
	Comment string `json:"column_comm"`
}

type QueryList struct {
	Queries []string `json:"queries"`
}

type EmptyResponse struct {
	Message string `json:"msg"`
}
