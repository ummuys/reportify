package convert

import "os"

type ReportConvert interface {
	ToPDF(headers []string, data [][]any, f *os.File) error
	// ToXLXS()
	// ToJSON()
	// ToCSV()
	// ToMD()

	//Maybe
	//ToTSV()
	//ToYAML()
	//ToParquet()
	//ToArrow()
	//ToSQLDump()

	//Maybe maybe
	//ToSuperSet()
}
