package auth

type Identity struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	CraftsmanID string `json:"craftsman_id,omitempty"`
}
