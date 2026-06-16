package ship

type ShipOrderRequest struct {
	OrderID uint64 `json:"order_id"`
	CraftsmanID uint64 `json:"craftsman_id"`
	CustomerID uint64 `json:"customer_id"`
}