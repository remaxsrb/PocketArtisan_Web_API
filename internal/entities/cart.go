package entities

type CartItem struct {
	ID uint64 `json:"id" gorm:"primaryKey;autoIncrement"`

	CartID    uint64 `json:"cart_id" gorm:"not null;uniqueIndex:idx_cart_product"`
	ProductID uint64 `json:"product_id" gorm:"not null;uniqueIndex:idx_cart_product"`

	Quantity int     `json:"quantity" gorm:"not null;default:1"`
	Product  Product `json:"product" gorm:"foreignKey:ProductID"`
	Cart     *Cart   `json:"-" gorm:"foreignKey:CartID"`
}

type Cart struct {
	ID     uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID uint64     `json:"user_id" gorm:"uniqueIndex;not null"`
	Items  []CartItem `json:"items,omitempty" gorm:"foreignKey:CartID;references:ID;constraint:OnDelete:CASCADE"`
}
