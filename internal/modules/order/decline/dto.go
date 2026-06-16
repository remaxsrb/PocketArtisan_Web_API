package decline

type DeclineOrderRequest struct {
	OrderID string `json:"order_id"`
	CraftsmanID string `json:"craftsman_id"`
}
