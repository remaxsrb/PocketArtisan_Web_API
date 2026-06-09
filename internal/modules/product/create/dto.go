package create

type NewProductRequest struct {
	Name          string  `json:"name"`
	CraftsmanID   uint64  `json:"craftsmanId"`
	MaterialPrice float64 `json:"materialPrice"`
	LaborPrice    float64 `json:"laborPrice"`
	Picture       string  `json:"picture"`
	Description   string  `json:"description"`
}

