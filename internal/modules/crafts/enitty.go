package crafts

type Craft struct {
	ID   int    `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Name string `json:"name" gorm:"not null; unique"`
}
