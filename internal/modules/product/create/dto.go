package create

type NewProductRequest struct {
	Name          string  `json:"name"`
	MaterialPrice float64 `json:"materialPrice"`
	LaborPrice    float64 `json:"laborPrice"`
	Picture       string  `json:"picture"`
}
