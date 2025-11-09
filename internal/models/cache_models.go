package models

import "time"

type CacheValue struct {
	ReportName string
	ReportComm string
	CreatedAt  time.Time
	Sql        string
	CSVSep     rune
}
