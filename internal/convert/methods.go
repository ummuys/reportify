package convert

import (
	"fmt"
	"os"

	"github.com/phpdave11/gofpdf"
	"github.com/rs/zerolog"
)

type repConv struct {
	logger *zerolog.Logger
}

func NewReportConvert(logger *zerolog.Logger) ReportConvert {
	return &repConv{logger: logger}
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
		{"A4", "P", 10, 12}, // ← пробуем сначала портретный A4
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
		pdf.SetMargins(p.leftRight, p.topBottom, p.leftRight)
		pdf.SetAutoPageBreak(true, p.topBottom)
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
		hadToShrink  bool    // понадобилось ли сжатие/уменьшение кегля на этапе baseFontSize
		shrinkFactor float64 // итоговое сжатие (<1) относительно исходной ширины на выбранном кегле
	}

	tryFit := func(p paperPreset) fitResult {
		pdf := newPDF(p)
		pageW, _ := pdf.GetPageSize()
		usableW := pageW - 2*p.leftRight

		// 1) проверяем на БАЗОВОМ кегле: влезает без сжатия?
		cwBase, sumBase := measure(pdf, baseFontSize)
		if sumBase <= usableW {
			// растягиваем, чтобы занять всю ширину
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

		// 2) иначе — пытаемся уменьшить кегль и/или сжать
		font := baseFontSize
		cw := cwBase
		sum := sumBase
		hadToShrink := true

		for {
			// если не влазит — сначала уменьшаем кегль
			if sum > usableW && font > minFontSize {
				font -= 0.5
				cw, sum = measure(pdf, font)
				continue
			}
			// если всё ещё не влазит — сжимаем пропорционально
			if sum > usableW {
				scale := usableW / sum
				for i := range cw {
					cw[i] *= scale
					if cw[i] < minColWidthMM {
						cw[i] = minColWidthMM
					}
				}
				// пересчёт суммы
				sum = 0
				for _, v := range cw {
					sum += v
				}
				// финальная нормализация
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

		// растягиваем до полной ширины (чтобы не было пустот)
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
			shrinkFactor: usableW / sumBase, // насколько пришлось ужимать от базовой ширины
		}
	}

	// ——— выбор лучшего пресета ———

	var chosen fitResult
	chosenSet := false

	// 1) сначала ищем первый пресет, где НЕ пришлось сжимать на baseFontSize (т.е. A4-P победит, если реально влазит)
	for _, p := range presets {
		fr := tryFit(p)
		if !fr.hadToShrink {
			chosen = fr
			chosenSet = true
			break
		}
	}

	// 2) если нигде "без сжатия" не получилось — берём тот, где минимальная потеря качества:
	// максимальный итоговый кегль и наименьшее сжатие
	if !chosenSet {
		var bestFont float64 = -1
		var bestShrink float64 = -1
		var best fitResult

		for _, p := range presets {
			fr := tryFit(p)
			scoreFont := fr.fontSize
			scoreShrink := fr.shrinkFactor // ближе к 1 — лучше

			if scoreFont > bestFont || (scoreFont == bestFont && scoreShrink > bestShrink) {
				bestFont = scoreFont
				bestShrink = scoreShrink
				best = fr
			}
		}
		chosen = best
	}

	pdf := chosen.pdf
	pdf.AddPage()
	pdf.SetFont("DejaVu", "", chosen.fontSize)

	// заголовок таблицы
	pdf.SetFillColor(240, 240, 240)
	pdf.SetDrawColor(200, 200, 200)
	for i, htxt := range headers {
		pdf.CellFormat(chosen.colW[i], headerH, htxt, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// строки
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
		rc.logger.Debug().Str("msg", "err to create PDF").Msg("")
		return err
	}
	rc.logger.Debug().Str("msg", "successful create PDF").Msg("")
	return nil
}
