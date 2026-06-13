package cart

type CartItemResponse struct {
	ProductID uint64 `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
