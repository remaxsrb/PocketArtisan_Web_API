package set_biography

type SetBiographyRequest struct {
	CraftsmanID uint64 `json:"craftsmanId" binding:"required"`
	Biography   string `json:"biography" binding:"max=200"`
}