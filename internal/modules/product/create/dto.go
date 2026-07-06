package create

type NewProductRequest struct {
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
	Videos      []string `json:"videos"`
	Category    string   `json:"category"`
	CraftsmanID uint64   `json:"-"`
}
