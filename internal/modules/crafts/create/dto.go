package create

type NewCraftRequest struct {
	Name     string   `json:"name" binding:"required"`
	Keywords []string `json:"keywords" binding:"required"`
}

type GetCraftsRequest struct {
	Crafts []string `json:"crafts"`
}
