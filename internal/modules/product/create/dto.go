package create

type NewProductRequest struct {
	Name          string   `json:"name"`
	CraftsmanID   uint64   `json:"craftsmanId"`
	MaterialPrice float64  `json:"materialPrice"`
	LaborPrice    float64  `json:"laborPrice"`
	Description   string   `json:"description"`
	Images        []string `json:"images"`
	Videos        []string `json:"videos"`
}
