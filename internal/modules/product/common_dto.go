package product

type ProductResponse struct {
	ID              uint64   `json:"id"`
	CraftsmanID uint64  `json:"craftsmanId" gorm:"column:craftsman_id"`
	Name            string   `json:"name"`
	Hidden          bool     `json:"hidden"`
	Price           float64  `json:"price"`
	Description     string   `json:"description"`
	Rating          float64  `json:"rating"`
	NumberOfRatings int      `json:"numberOfRatings"`
	Available       bool     `json:"available"`
	Images          []string `json:"images"`
	Videos          []string `json:"videos"`
}
