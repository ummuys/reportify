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

		logoPath  = "internal/convert/pgups_icon.png"
		logoWmm   = 18.0 // ширина логотипа (высота сохранит пропорции)
		logoTopY  = 6.0  // отступ логотипа от верхнего края страницы
		logoSpace = 14.0 // дополнительное место над контентом под логотип
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
		rc.logger.Debug().Str("msg", "err to create PDF").Msg("")
		return err
	}
	rc.logger.Debug().Str("msg", "successful create PDF").Msg("")
	return nil
}
