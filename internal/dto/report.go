package dto

import (
	"os"

	"github.com/ummuys/reportify/internal/webdto"
)

type CreateReport struct {
	UserID       int64
	ReportParams webdto.ReportParams
	ReportFile   *os.File
	Format       string
}
