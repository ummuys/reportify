package convert

import "os"

type ReportConvert interface {
	ToPDF(headers []string, data [][]any, f *os.File) error
	ToXLSX(headers []string, data [][]any, f *os.File) error
	ToJSON(headers []string, data [][]any, f *os.File) error
	ToCSV(headers []string, data [][]any, f *os.File, sep rune) error
	ToDOCX(headers []string, data [][]any, f *os.File) error

	// ToMD()

	// Maybe
	// ToTSV()
	// ToYAML()
	// ToParquet()
	// ToArrow()
	// ToSQLDump()

	// Maybe maybe
	// ToSuperSet()
}
