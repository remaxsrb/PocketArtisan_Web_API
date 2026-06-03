package product

type Product struct {
	ID     uint64  `json:"id" gorm:"primary_key"`
	Price  float64 `json:"price" gorm:"not null"`
	Name   string  `json:"name" gorm:"not null"`
	Hidden bool    `json:"hidden" gorm:"not null"`
}
