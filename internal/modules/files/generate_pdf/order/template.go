package order

import (
	"PocketArtisan/internal/modules/utils/fonts"
	"fmt"

	"github.com/go-pdf/fpdf"
)

const (
	pageW    = 210.0
	margin   = 15.0
	contentW = pageW - 2*margin
	rowH     = 7.0
)

func buildPDF(data OrderData, f *fonts.Service) (*fpdf.Fpdf, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(margin, margin, margin)

	pdf.AddUTF8FontFromBytes("DejaVuSans", "", f.Regular)
	pdf.AddUTF8FontFromBytes("DejaVuSans", "B", f.Bold)
	pdf.AddUTF8FontFromBytes("DejaVuSans", "I", f.Italic)

	pdf.AddPage()

	// -- Header ---------------------------------------------------------------
	pdf.SetFillColor(30, 30, 60)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("DejaVuSans", "B", 20)
	pdf.CellFormat(contentW, 14, "Фактура", "", 1, "C", true, 0, "")
	pdf.Ln(4)

	// -- Order meta -----------------------------------------------------------
	pdf.SetTextColor(40, 40, 40)
	metaLabel := func(label, value string) {
		pdf.SetFont("DejaVuSans", "B", 10)
		pdf.Cell(40, rowH, label+":")
		pdf.SetFont("DejaVuSans", "", 10)
		pdf.Cell(contentW-40, rowH, value)
		pdf.Ln(rowH)
	}
	metaLabel("ID Porudžbine", fmt.Sprintf("#%d", data.OrderID))
	metaLabel("Датум", data.OrderDate)
	metaLabel("Купац", data.CustomerName)
	metaLabel("E-пошта", data.CustomerEmail)
	metaLabel("Адреса доставе", data.ShippingAddress)
	metaLabel("Начин плаћања", data.PaymentType)
	pdf.Ln(4)

	// -- Items table header ---------------------------------------------------
	colProduct := contentW * 0.50
	colQty := contentW * 0.15
	colUnit := contentW * 0.175
	colTotal := contentW * 0.175

	pdf.SetFillColor(50, 50, 90)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("DejaVuSans", "B", 10)
	pdf.CellFormat(colProduct, rowH, "Производ", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colQty, rowH, "Количина", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colUnit, rowH, "Цена", "1", 0, "R", true, 0, "")
	pdf.CellFormat(colTotal, rowH, "Укупно", "1", 1, "R", true, 0, "")

	// -- Item rows ------------------------------------------------------------
	pdf.SetTextColor(40, 40, 40)
	pdf.SetFont("DejaVuSans", "", 10)
	fill := false
	for _, item := range data.Items {
		if fill {
			pdf.SetFillColor(235, 235, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		lineTotal := item.UnitPrice * float64(item.Quantity)
		pdf.CellFormat(colProduct, rowH, item.Product.Name, "1", 0, "L", true, 0, "")
		pdf.CellFormat(colQty, rowH, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", true, 0, "")
		pdf.CellFormat(colUnit, rowH, fmt.Sprintf("%.2f РСД", item.UnitPrice), "1", 0, "R", true, 0, "")
		pdf.CellFormat(colTotal, rowH, fmt.Sprintf("%.2f РСД", lineTotal), "1", 1, "R", true, 0, "")
		fill = !fill
	}

	// -- Total row ------------------------------------------------------------
	pdf.SetFillColor(30, 30, 60)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("DejaVuSans", "B", 11)
	pdf.CellFormat(colProduct+colQty+colUnit, rowH, "Укупно", "1", 0, "R", true, 0, "")
	pdf.CellFormat(colTotal, rowH, fmt.Sprintf("%.2f РСД", data.TotalPrice), "1", 1, "R", true, 0, "")

	// -- Footer ---------------------------------------------------------------
	pdf.Ln(8)
	pdf.SetTextColor(120, 120, 120)
	pdf.SetFont("DejaVuSans", "I", 9)
	pdf.MultiCell(contentW, 5, "Хвала Вам на поруџбини!", "", "C", false)

	return pdf, nil
}
