package webdto

import "time"

type RawReportParams struct {
	ReportName string    `json:"report_name"`
	ReportComm string    `json:"report_comm"`
	CreatedAt  time.Time `json:"created_at"`
	Sql        string    `json:"sql"`
	CSVSep     string    `json:"csv_sep"`
}

type ReportParams struct {
	ReportName string
	ReportComm string
	CreatedAt  time.Time
	Sql        string
	CSVSep     rune
}
