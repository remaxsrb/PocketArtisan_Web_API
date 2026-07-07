package addtocart

type AddToCartRequest struct {
	ProductID uint64 `json:"product_id"`
	Quantity  int    `json:"quantity"`
	UserID    uint64 `json:"-"`
}
