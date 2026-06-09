package product

type ProductResponse struct {
	ID          uint64         `json:"id"`
	Name        string         `json:"name"`
	CraftsmanID uint64         `json:"craftsmanId"`
	Hidden      bool           `json:"hidden"`
	TotalPrice  float64        `json:"totalPrice"`
	Description string         `json:"description"`
	Images      []ProductImage `json:"images"`
	Videos      []ProductVideo `json:"videos"`
}
