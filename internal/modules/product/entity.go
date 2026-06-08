package product

type Product struct {
	ID            uint64  `json:"id" gorm:"primary_key"`
	Name          string  `json:"name" gorm:"not null"`
	Hidden        bool    `json:"hidden" gorm:"not null"`
	Picture       string  `json:"picture" gorm:"not null"`
	MaterialPrice float64 `json:"materialPrice" gorm:"not null"`
	LaborPrice    float64 `json:"laborPrice" gorm:"not null"`
}
