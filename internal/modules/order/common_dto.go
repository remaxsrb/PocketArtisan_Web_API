package order

import (
	"PocketArtisan/internal/entities"
	"time"
)
type OrderResponse struct {
	OrderID         uint64  `json:"order_id"`
	CustomerID	  uint64  `json:"customer_id,omitempty"`
	CraftsmanID	  uint64  `json:"craftsman_id,omitempty"`
	OrderDate       time.Time  `json:"order_date"` // dd/mm/yyyy
	CompletionDate  time.Time  `json:"completion_date,omitempty"`
	URL			string  `json:"url,omitempty"`
	Status entities.OrderStatus `json:"status"`
}