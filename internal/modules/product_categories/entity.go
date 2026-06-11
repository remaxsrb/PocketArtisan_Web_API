package product_categories

type ProductCategory struct {
	ID   int    `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Name string `json:"name" gorm:"not null; unique"`
}
