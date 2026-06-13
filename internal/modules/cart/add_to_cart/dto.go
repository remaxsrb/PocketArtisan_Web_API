package addtocart

import "PocketArtisan/internal/entities"

type AddToCartRequest struct {
	CartID    uint64 `json:"cart_id"`
	ProductID uint64 `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type AddToCartResponse struct {
	Cart entities.Cart `json:"cart"`
}
