package convert

import (
	"fmt"
	"os"

	"github.com/phpdave11/gofpdf"
	"github.com/rs/zerolog"
)

const (
	marginL  = 10.0
	marginR  = 10.0
	rowPad   = 2.0
	headerH  = 8.0
	baseH    = 6.0
	fontName = "DejaVu"
	fontSize = 10.0
	headerSz = 11.0
)

//TODO:
// 👉 В результате получится таблица, где:

// заголовки читаемы и разделены;

// данные не накладываются и не обрезаются;

// кириллица/UTF-8 отображается нормально;

// числа и даты форматированы.
type repConv struct {
	logger *zerolog.Logger
}

func NewReportConvert(logger *zerolog.Logger) ReportConvert {
	return &repConv{logger: logger}
}

func (rc *repConv) ToPDF(headers []string, rows [][]any, f *os.File) error {
	rc.logger.Debug().Str("env", "call toPDF").Msg("")

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 10)
	pdf.AddPage()

	cellW := 190.0 / float64(len(headers))

	for _, h := range headers {
		pdf.CellFormat(cellW, 8, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	for _, row := range rows {
		for _, v := range row {
			pdf.CellFormat(cellW, 6, fmt.Sprint(v), "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	if err := pdf.Output(f); err != nil {
		rc.logger.Debug().Str("msg", "err to create PDF").Msg("")
		return err
	}
	rc.logger.Debug().Str("msg", "successful create PDF").Msg("")
	return nil
}
