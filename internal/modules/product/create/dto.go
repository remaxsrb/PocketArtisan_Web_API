package create

type NewProductRequest struct {
	Name        string   `json:"name"`
	CraftsmanID uint64   `json:"craftsmanId"`
	Price       float64  `json:"price"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
	Videos      []string `json:"videos"`
}
