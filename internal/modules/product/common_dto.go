package product

type ProductResponse struct {
	Name        string  `json:"name"`
	CraftsmanID uint64  `json:"craftsmanId"`
	Hidden      bool    `json:"hidden"`
	Picture     string  `json:"picture"`
	TotalPrice  float64 `json:"totalPrice"`
	Description string  `json:"description"`
}