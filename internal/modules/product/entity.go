package product

import (
	"PocketArtisan/internal/modules/product_categories"
)

type Product struct {
	ID              uint64  `json:"id" gorm:"primaryKey;autoIncrement"`
	CraftsmanID     uint64  `json:"craftsman_id" gorm:"not null;uniqueIndex:idx_craftsman_product"`
	Name            string  `json:"name" gorm:"not null;uniqueIndex:idx_craftsman_product"`
	Hidden          bool    `json:"hidden" gorm:"not null"`
	Price           float64 `json:"price" gorm:"not null"`
	Description     string  `json:"description" gorm:"not null"`
	Rating          float64 `json:"rating" gorm:"not null;default:0.0"`
	NumberOfRatings int     `json:"number_of_ratings" gorm:"not null;default:0"`
	Available       bool    `json:"available" gorm:"not null;default:true"`
	CategoryID      uint64  `json:"category_id" gorm:"not null"`

	Images   []ProductImage                      `json:"images" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	Videos   []ProductVideo                      `json:"videos" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	Category *product_categories.ProductCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID;constraint:OnDelete:RESTRICT;"`
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
