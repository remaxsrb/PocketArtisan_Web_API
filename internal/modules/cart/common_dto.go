package cart

import "PocketArtisan/internal/entities"

type CartItemResponse struct {
	ProductID uint64 `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CartResponse struct {
	Cart entities.Cart `json:"cart"`
}
