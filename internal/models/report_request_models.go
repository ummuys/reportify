package models

type RawReportParams struct {
	Sql    string `json:"sql"`
	CSVSep string `json:"csv_sep"`
}

type ReportParams struct {
	Sql    string
	CSVSep rune
}
