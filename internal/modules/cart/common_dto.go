package cart

type CartItemResponse struct {
	ID        uint64 `json:"id"`
	ProductID uint64 `json:"product_id"`
	Quantity  int    `json:"quantity"`
}