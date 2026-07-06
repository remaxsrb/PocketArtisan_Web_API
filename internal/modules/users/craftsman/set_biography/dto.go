package set_biography

type SetBiographyRequest struct {
	CraftsmanID uint64
	Biography   string `json:"biography" binding:"max=200"`
}