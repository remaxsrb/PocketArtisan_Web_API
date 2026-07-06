package toggle_hide

type ToggleHideProduct struct {
	ProductID   uint64 `json:"product_id"`
	CraftsmanID uint64 `json:"-"`
}
