package addtocart

type AddToCartRequest struct {
	CartID    uint64 `json:"cart_id"`
	ProductID uint64 `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
