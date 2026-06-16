package accept


type AcceptOrderRequest struct {
	OrderID uint64 `json:"order_id"`
	CraftsmanID uint64 `json:"craftsman_id"`
}
