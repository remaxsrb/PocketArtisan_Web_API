package product

type Product struct {
	Name          string  `json:"name" gorm:"primaryKey;autoIncrement:false;not null"`
    CraftsmanID   uint64  `json:"craftsman_id" gorm:"primaryKey;autoIncrement:false;not null"`
	Hidden        bool    `json:"hidden" gorm:"not null"`
	Picture       string  `json:"picture" gorm:"not null"`
	MaterialPrice float64 `json:"materialPrice" gorm:"not null"`
	LaborPrice    float64 `json:"laborPrice" gorm:"not null"`
	Description   string  `json:"description" gorm:"not null"`
}

/*


Idea is to have a composite primary key of (name, craftsman_id) to allow different craftsmen to have products with the same name, 
while preventing a single craftsman from having multiple products with the same name.

*/
