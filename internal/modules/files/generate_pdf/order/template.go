package order

import (
	"fmt"

	"github.com/go-pdf/fpdf"
)

const (
	pageW      = 210.0
	margin     = 15.0
	contentW   = pageW - 2*margin
	rowH       = 7.0
	headerBlue = 0x1A // R component for a dark navy
)

func buildPDF(data OrderData) (*fpdf.Fpdf, error) {
	f := fpdf.New("P", "mm", "A4", "")
	f.SetMargins(margin, margin, margin)
	f.AddPage()

	// ── Header ──────────────────────────────────────────────────────────────
	f.SetFillColor(30, 30, 60)
	f.SetTextColor(255, 255, 255)
	f.SetFont("Helvetica", "B", 20)
	f.CellFormat(contentW, 14, "Order Confirmation", "", 1, "C", true, 0, "")
	f.Ln(4)

	// ── Order meta ──────────────────────────────────────────────────────────
	f.SetTextColor(40, 40, 40)
	f.SetFont("Helvetica", "", 10)
	metaLabel := func(label, value string) {
		f.SetFont("Helvetica", "B", 10)
		f.Cell(40, rowH, label+":")
		f.SetFont("Helvetica", "", 10)
		f.Cell(contentW-40, rowH, value)
		f.Ln(rowH)
	}
	metaLabel("Order ID", fmt.Sprintf("#%d", data.OrderID))
	metaLabel("Date", data.OrderDate)
	metaLabel("Customer", data.CustomerName)
	metaLabel("Email", data.CustomerEmail)
	metaLabel("Shipping Address", data.ShippingAddress)
	metaLabel("Payment", data.PaymentType)
	f.Ln(4)

	// ── Items table header ───────────────────────────────────────────────────
	colProduct := contentW * 0.50
	colQty := contentW * 0.15
	colUnit := contentW * 0.175
	colTotal := contentW * 0.175

	f.SetFillColor(50, 50, 90)
	f.SetTextColor(255, 255, 255)
	f.SetFont("Helvetica", "B", 10)
	f.CellFormat(colProduct, rowH, "Product", "1", 0, "L", true, 0, "")
	f.CellFormat(colQty, rowH, "Qty", "1", 0, "C", true, 0, "")
	f.CellFormat(colUnit, rowH, "Unit Price", "1", 0, "R", true, 0, "")
	f.CellFormat(colTotal, rowH, "Subtotal", "1", 1, "R", true, 0, "")

	// ── Item rows ────────────────────────────────────────────────────────────
	f.SetTextColor(40, 40, 40)
	f.SetFont("Helvetica", "", 10)
	fill := false
	for _, item := range data.Items {
		if fill {
			f.SetFillColor(235, 235, 245)
		} else {
			f.SetFillColor(255, 255, 255)
		}
		subtotal := item.UnitPrice * float64(item.Quantity)
		f.CellFormat(colProduct, rowH, item.Product.Name, "1", 0, "L", true, 0, "")
		f.CellFormat(colQty, rowH, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", true, 0, "")
		f.CellFormat(colUnit, rowH, fmt.Sprintf("$%.2f", item.UnitPrice), "1", 0, "R", true, 0, "")
		f.CellFormat(colTotal, rowH, fmt.Sprintf("$%.2f", subtotal), "1", 1, "R", true, 0, "")
		fill = !fill
	}

	// ── Total row ────────────────────────────────────────────────────────────
	f.SetFillColor(30, 30, 60)
	f.SetTextColor(255, 255, 255)
	f.SetFont("Helvetica", "B", 11)
	f.CellFormat(colProduct+colQty+colUnit, rowH, "Total", "1", 0, "R", true, 0, "")
	f.CellFormat(colTotal, rowH, fmt.Sprintf("$%.2f", data.TotalPrice), "1", 1, "R", true, 0, "")

	// ── Footer ───────────────────────────────────────────────────────────────
	f.Ln(8)
	f.SetTextColor(120, 120, 120)
	f.SetFont("Helvetica", "I", 9)
	f.MultiCell(contentW, 5, "Thank you for your order! A craftsman will review it shortly.", "", "C", false)

	return f, nil
}
