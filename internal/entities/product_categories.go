package entities

import "github.com/lib/pq"

type ProductCategory struct {
	ID             uint64         `json:"id" gorm:"primaryKey;autoIncrement:true"`
	Name           string         `json:"name" gorm:"not null; unique"`
	Keywords       pq.StringArray `gorm:"type:text[]"`
	SearchKeywords pq.StringArray `gorm:"type:text[];index:,type:gin"`
}
