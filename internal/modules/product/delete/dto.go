package delete

type DeleteProductRequest struct {
	ProductID uint64 `json:"id"`
	CraftsmanID uint64 `json:"craftsmanId"`
}