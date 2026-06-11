package get_all

type NewCraftRequest struct {
	Name string `json:"name" gorm:"not null;"`
}

type GetAllCraftsRequest struct {
	Crafts []string `json:"crafts"`
}
