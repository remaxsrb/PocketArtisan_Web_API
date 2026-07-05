package entities

type CraftProductCategory struct {
	CraftID    uint64 `json:"craft_id" gorm:"primaryKey"`
	CategoryID uint64 `json:"category_id" gorm:"primaryKey"`

	Craft    *Craft           `json:"craft,omitempty" gorm:"foreignKey:CraftID;constraint:OnDelete:CASCADE;"`
	Category *ProductCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE;"`
}