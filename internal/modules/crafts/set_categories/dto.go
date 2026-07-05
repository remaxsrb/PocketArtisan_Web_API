package setcategories

type SetCraftCategoriesRequest struct {
	Craft      string   `json:"craft" binding:"required"`
	Categories []string `json:"categories" binding:"required"`
}