package order

import "PocketArtisan/internal/entities"

type OrderData struct {
	OrderID         uint64               `json:"order_id"`
	CustomerName    string               `json:"customer_name"`
	CustomerEmail   string               `json:"customer_email"`
	ShippingAddress string               `json:"shipping_address"`
	PaymentType     string               `json:"payment_type"`
	Items           []entities.OrderItem `json:"items"`
	OrderDate       string               `json:"order_date"` // dd/mm/yyyy
	TotalPrice      float64              `json:"total_price"`
}
