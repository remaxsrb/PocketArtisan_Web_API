package product

type ProductResponse struct {
	ID          uint64   `json:"id"`
	Name        string   `json:"name"`
	Hidden      bool     `json:"hidden"`
	Price       float64  `json:"price"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
	Videos      []string `json:"videos"`
}
