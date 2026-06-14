package removefromcart

type RemoveFromCartRequest struct {
	CartID    uint64 `json:"cart_id"`
	ProductID uint64 `json:"product_id"`
	Quantity  uint64 `json:"quantity"`
}
