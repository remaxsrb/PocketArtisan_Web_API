package removefromcart

type RemoveFromCartRequest struct {
	ProductID uint64 `json:"product_id"`
	UserID    uint64 `json:"-"`
}
