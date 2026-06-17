package checkout

import "PocketArtisan/internal/entities"

type CheckoutRequest struct {
	PaymentType     entities.PaymentType `json:"payment_type" binding:"required"`
	ShippingAddress string               `json:"shipping_address" binding:"required"`
}

type OrderResult struct {
	OrderID     uint64  `json:"order_id"`
	CraftsmanID uint64  `json:"craftsman_id"`
	Total       float64 `json:"total"`
	PDFURL      string  `json:"pdf_url,omitempty"`
}