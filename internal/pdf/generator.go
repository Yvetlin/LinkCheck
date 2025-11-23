package pdf

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/jung-kurt/gofpdf"
	"github.com/linkcheck/pkg/types"
)

// Generator генерирует PDF отчеты о статусе ссылок
type Generator struct{}

// NewGenerator создает новый генератор PDF
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateReport создает PDF отчет для указанных наборов ссылок
func (g *Generator) GenerateReport(linksSets map[int]types.LinksSet) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Link Status Report")
	pdf.Ln(12)

	var linksNums []int
	for num := range linksSets {
		linksNums = append(linksNums, num)
	}
	sort.Ints(linksNums)

	pdf.SetFont("Arial", "", 12)

	for _, linksNum := range linksNums {
		set := linksSets[linksNum]

		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, fmt.Sprintf("Links Set #%d", linksNum))
		pdf.Ln(8)

		var urls []string
		for url := range set.Links {
			urls = append(urls, url)
		}
		sort.Strings(urls)

		pdf.SetFont("Arial", "", 11)
		for _, url := range urls {
			status := set.Links[url]

			statusText := "a"
			if status == types.StatusNotAvailable {
				statusText = "na"
			}

			line := fmt.Sprintf("%s - %s", statusText, url)

			pdf.MultiCell(0, 7, line, "", "", false)
		}

		pdf.Ln(5)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// FormatStatus преобразует статус в короткий текст
func (g *Generator) FormatStatus(status types.LinkStatus) string {
	if status == types.StatusAvailable {
		return "a"
	}
	return "na"
}
