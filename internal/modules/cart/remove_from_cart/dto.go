package removefromcart

import "PocketArtisan/internal/modules/cart"

type RemoveFromCartRequest struct {
	UserID    uint64 `json:"user_id"`
	ProductID uint64 `json:"product_id"`
}

type RemoveFromCartResponse struct {
	CartItems []cart.CartItemResponse `json:"cart_items"`
}
