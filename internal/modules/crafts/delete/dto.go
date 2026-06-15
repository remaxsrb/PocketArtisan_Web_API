package delete

type DeleteCraftRequest struct {
	Name string `json:"name" binding:"required"`
}
