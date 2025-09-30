package models

type ListSchemas struct {
	Schemas []string `json:"schemas"`
}

type ListTables struct {
	Tables []string `json:"tables"`
}

type ListColumns struct {
	Columns []string `json:"columns"`
}
