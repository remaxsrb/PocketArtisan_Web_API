package addtocart

import "PocketArtisan/internal/modules/cart"

type AddToCartRequest struct {
	UserID    uint64 `json:"user_id"`
	ProductID uint64 `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type AddToCartResponse struct {
	CartItems []cart.CartItemResponse `json:"cart_items"`
}

