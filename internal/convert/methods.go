package convert

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"baliance.com/gooxml/color"
	"baliance.com/gooxml/document"
	"baliance.com/gooxml/measurement"
	"baliance.com/gooxml/schema/soo/wml"
	"github.com/phpdave11/gofpdf"
	"github.com/rs/zerolog"
	"github.com/xuri/excelize/v2"
)

type repConv struct {
	logger *zerolog.Logger
}

func NewReportConvert(logger *zerolog.Logger, SSUser bool) ReportConvert {
	return &repConv{logger: logger}
}

func (rc *repConv) ToDOCX(headers []string, data [][]any, f *os.File) error {
	rc.logger.Debug().Str("env", "call toDOCX").Msg("")

	doc := document.New()

	table := doc.AddTable()
	table.Properties().SetWidthPercent(100)
	tblBorders := table.Properties().Borders()
	th := measurement.Distance(0.5 * measurement.Point) // толщина линии ~0.5pt
	tblBorders.SetAll(wml.ST_BorderSingle, color.Auto, th)
	tblBorders.SetInsideVertical(wml.ST_BorderSingle, color.Auto, th)
	tblBorders.SetInsideHorizontal(wml.ST_BorderSingle, color.Auto, th)

	header := table.AddRow()
	for _, h := range headers {
		cell := header.AddCell()
		para := cell.AddParagraph()
		run := para.AddRun()
		run.AddText(h)
		run.Properties().SetBold(true)
	}

	for _, rowData := range data {
		row := table.AddRow()
		for _, d := range rowData {
			cell := row.AddCell()
			cell.AddParagraph().AddRun().AddText(fmt.Sprint(d))
		}
	}

	if err := doc.Save(f); err != nil {
		rc.logger.Error().Err(err).Msg("fatal create DOCX")
		return fmt.Errorf("can't save in DOCX file: %v", err)
	}

	return nil
}

func (rc *repConv) ToJSON(headers []string, data [][]any, f *os.File) error {
	rc.logger.Debug().Str("env", "call toJSON").Msg("")

	res := make([]map[string]any, len(data))
	for i, row := range data {
		m := make(map[string]any)
		for j, d := range row {
			m[headers[j]] = d
		}
		res[i] = m
	}

	if err := json.NewEncoder(f).Encode(res); err != nil {
		rc.logger.Error().Err(err).Msg("fatal create JSON")
		return fmt.Errorf("can't save in JSON file: %v", err)
	}

	return nil
}

func (rc *repConv) ToXLSX(headers []string, data [][]any, f *os.File) error {
	rc.logger.Debug().Str("env", "call toXLSX").Msg("")

	fx := excelize.NewFile()
	defer func() {
		if err := fx.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheet := "Sheet1"

	idx, err := fx.NewSheet(sheet)
	if err != nil {
		return err
	}

	_, err = fx.GetSheetIndex(sheet)
	if err != nil {
		return fmt.Errorf("get new sheet: %w", err)
	}

	for col, head := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return fmt.Errorf("can't conv int -> cell: %v", err)
		}
		if err := fx.SetCellValue(sheet, cell, head); err != nil {
			return err
		}
	}

	for row, record := range data {
		for col, val := range record {
			cell, err := excelize.CoordinatesToCellName(col+1, row+2)
			if err != nil {
				return fmt.Errorf("can't conv int -> cell: %v", err)
			}
			if err := fx.SetCellValue(sheet, cell, val); err != nil {
				return err
			}
		}
	}

	fx.SetActiveSheet(idx)

	if err := fx.Write(f); err != nil {
		rc.logger.Error().Err(err).Msg("fatal create XLSX")
		return fmt.Errorf("can't save in XLSX file: %v", err)
	}

	rc.logger.Debug().Str("msg", "successful create XLSX").Msg("")
	return nil
}

func (rc *repConv) ToCSV(headers []string, rows [][]any, f *os.File, sep rune) error {
	rc.logger.Debug().Str("env", "call toCSV").Msg("")
	if len(headers) == 0 {
		return fmt.Errorf("empty headers")
	}

	w := csv.NewWriter(f)
	if sep != ' ' {
		w.Comma = sep
	}
	defer w.Flush()

	if err := w.Write(headers); err != nil {
		return fmt.Errorf("write headers: %w", err)
	}

	for _, row := range rows {
		strs := make([]string, len(row))
		for i, r := range row {
			strs[i] = fmt.Sprint(r)
		}
		if err := w.Write(strs); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}

	if err := w.Error(); err != nil {
		rc.logger.Debug().Str("msg", "fatal create CSV").Msg("")
		return fmt.Errorf("csv writer error: %w", err)
	}

	rc.logger.Debug().Str("msg", "successful create CSV").Msg("")
	return nil
}

func (rc *repConv) ToPDF(headers []string, rows [][]any, f *os.File) error {
	rc.logger.Debug().Str("env", "call toPDF").Msg("")

	if len(headers) == 0 {
		return fmt.Errorf("empty headers")
	}

	const (
		baseFontSize  = 10.0
		minFontSize   = 6.5
		cellPad       = 1.5
		headerH       = 8.0
		rowH          = 6.0
		sampleRows    = 80
		minColWidthMM = 14.0
		maxColWidthMM = 70.0

		logoPath  = "internal/convert/pgups_icon.png"
		logoWmm   = 18.0 // ширина логотипа (высота сохранит пропорции)
		logoTopY  = 6.0  // отступ логотипа от верхнего края страницы
		logoSpace = 20   // дополнительное место над контентом под логотип
	)

	sampleN := sampleRows
	if len(rows) < sampleN {
		sampleN = len(rows)
	}

	ttf, err := os.ReadFile("internal/convert/fonts/DejaVuSans.ttf")
	if err != nil {
		return err
	}

	type paperPreset struct {
		sizeStr     string
		orientation string
		leftRight   float64
		topBottom   float64
	}
	presets := []paperPreset{
		{"A4", "P", 10, 12},
		{"A4", "L", 8, 10},
		{"A3", "P", 10, 12},
		{"A3", "L", 8, 10},
	}
	switch {
	case len(headers) >= 18:
		for i := range presets {
			presets[i].leftRight = 6
		}
	case len(headers) >= 12:
		for i := range presets {
			if presets[i].leftRight > 8 {
				presets[i].leftRight = 8
			}
		}
	}

	newPDF := func(p paperPreset) *gofpdf.Fpdf {
		pdf := gofpdf.New(p.orientation, "mm", p.sizeStr, "")
		pdf.AddUTF8FontFromBytes("DejaVu", "", ttf)
		pdf.SetFont("DejaVu", "", baseFontSize)

		// ВАЖНО: верхнее поле увеличиваем на высоту для логотипа
		pdf.SetMargins(p.leftRight, p.topBottom+logoSpace, p.leftRight)
		pdf.SetTopMargin(p.topBottom + logoSpace)
		pdf.SetAutoPageBreak(true, p.topBottom)

		// Регистрируем и рисуем логотип в хедере (позицию контента не трогаем)
		_ = pdf.RegisterImageOptions(logoPath, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true})
		pdf.SetHeaderFuncMode(func() {
			pageW, _ := pdf.GetPageSize()
			_, _, rm, _ := pdf.GetMargins()

			x := pageW - rm - logoWmm
			y := logoTopY
			pdf.ImageOptions(
				logoPath,
				x, y,
				logoWmm, 0, // высота по пропорциям
				false,
				gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
				0,
				"",
			)
		}, true)

		return pdf
	}

	measure := func(pdf *gofpdf.Fpdf, fontSize float64) ([]float64, float64) {
		pdf.SetFont("DejaVu", "", fontSize)
		colW := make([]float64, len(headers))
		for i, htxt := range headers {
			sw := pdf.GetStringWidth(htxt) + 2*cellPad
			if sw < minColWidthMM {
				sw = minColWidthMM
			}
			if sw > maxColWidthMM {
				sw = maxColWidthMM
			}
			colW[i] = sw
		}
		for r := 0; r < sampleN; r++ {
			row := rows[r]
			for c := 0; c < len(headers) && c < len(row); c++ {
				sw := pdf.GetStringWidth(fmt.Sprint(row[c])) + 2*cellPad
				if sw > colW[c] {
					if sw > maxColWidthMM {
						sw = maxColWidthMM
					}
					colW[c] = sw
				}
			}
		}
		sum := 0.0
		for _, w := range colW {
			sum += w
		}
		return colW, sum
	}

	type fitResult struct {
		pdf          *gofpdf.Fpdf
		colW         []float64
		usableW      float64
		fontSize     float64
		hadToShrink  bool
		shrinkFactor float64
	}

	tryFit := func(p paperPreset) fitResult {
		pdf := newPDF(p)
		pageW, _ := pdf.GetPageSize()
		usableW := pageW - 2*p.leftRight

		cwBase, sumBase := measure(pdf, baseFontSize)
		if sumBase <= usableW {
			grow := usableW / sumBase
			for i := range cwBase {
				cwBase[i] *= grow
			}
			return fitResult{
				pdf:          pdf,
				colW:         cwBase,
				usableW:      usableW,
				fontSize:     baseFontSize,
				hadToShrink:  false,
				shrinkFactor: 1.0,
			}
		}

		font := baseFontSize
		cw := cwBase
		sum := sumBase
		hadToShrink := true

		for {
			if sum > usableW && font > minFontSize {
				font -= 0.5
				cw, sum = measure(pdf, font)
				continue
			}
			if sum > usableW {
				scale := usableW / sum
				for i := range cw {
					cw[i] *= scale
					if cw[i] < minColWidthMM {
						cw[i] = minColWidthMM
					}
				}
				sum = 0
				for _, v := range cw {
					sum += v
				}
				if sum > usableW {
					ratio := usableW / sum
					for i := range cw {
						cw[i] *= ratio
					}
					sum = usableW
				}
				break
			}
			break
		}

		if sum < usableW && sum > 0 {
			grow := usableW / sum
			for i := range cw {
				cw[i] *= grow
			}
		}

		return fitResult{
			pdf:          pdf,
			colW:         cw,
			usableW:      usableW,
			fontSize:     font,
			hadToShrink:  hadToShrink,
			shrinkFactor: usableW / sumBase,
		}
	}

	var chosen fitResult
	chosenSet := false
	for _, p := range presets {
		fr := tryFit(p)
		if !fr.hadToShrink {
			chosen = fr
			chosenSet = true
			break
		}
	}
	if !chosenSet {
		var bestFont float64 = -1
		var bestShrink float64 = -1
		var best fitResult
		for _, p := range presets {
			fr := tryFit(p)
			if fr.fontSize > bestFont || (fr.fontSize == bestFont && fr.shrinkFactor > bestShrink) {
				bestFont = fr.fontSize
				bestShrink = fr.shrinkFactor
				best = fr
			}
		}
		chosen = best
	}

	pdf := chosen.pdf
	pdf.AddPage() // контент стартует ниже: topMargin уже увеличен на logoSpace
	pdf.SetFont("DejaVu", "", chosen.fontSize)

	pdf.SetFillColor(240, 240, 240)
	pdf.SetDrawColor(200, 200, 200)
	for i, htxt := range headers {
		pdf.CellFormat(chosen.colW[i], headerH, htxt, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	alt := false
	for _, row := range rows {
		alt = !alt
		if alt {
			pdf.SetFillColor(248, 248, 248)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		for i := range headers {
			var txt string
			if i < len(row) {
				txt = fmt.Sprint(row[i])
			}
			pdf.CellFormat(chosen.colW[i], rowH, txt, "1", 0, "L", true, 0, "")
		}
		pdf.Ln(-1)
	}

	if err := pdf.Output(f); err != nil {
		rc.logger.Error().Err(err).Msg("fatal create PDF")
		return err
	}
	rc.logger.Debug().Str("msg", "successful create PDF").Msg("")
	return nil
}
