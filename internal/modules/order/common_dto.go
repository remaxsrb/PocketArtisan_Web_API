package order

import (
	"PocketArtisan/internal/entities"
	"time"
)
type OrderResponse struct {
	OrderID         uint64  `json:"order_id"`
	OrderDate       time.Time  `json:"order_date"` // dd/mm/yyyy
	CompletionDate  time.Time  `json:"completion_date,omitempty"`
	URL			string  `json:"url,omitempty"`
	Status entities.OrderStatus `json:"status"`
}