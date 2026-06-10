package product

type Product struct {
	ID          uint64  `json:"id" gorm:"primaryKey;autoIncrement"`
	CraftsmanID uint64  `json:"craftsman_id" gorm:"not null;uniqueIndex:idx_craftsman_product"`
	Name        string  `json:"name" gorm:"not null;uniqueIndex:idx_craftsman_product"`
	Hidden      bool    `json:"hidden" gorm:"not null"`
	Price       float64 `json:"price" gorm:"not null"`
	Description string  `json:"description" gorm:"not null"`

	Images []ProductImage `json:"images" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	Videos []ProductVideo `json:"videos" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
}

type ProductImage struct {
	ID        uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID uint64 `json:"product_id" gorm:"not null;index"`
	URL       string `json:"url" gorm:"not null"`
}

type ProductVideo struct {
	ID        uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID uint64 `json:"product_id" gorm:"not null;index"`
	URL       string `json:"url" gorm:"not null"`
}
