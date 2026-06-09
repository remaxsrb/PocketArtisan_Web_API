package getbycraft

import "PocketArtisan/internal/modules/users"

type GetByCraftRequest struct {
	
	Limit int
	Skip  int
}

type GetByCraftResponse struct {
	Craftsmen []*users.CraftsmanResponse `json:"craftsmen"`
	Total     int64                      `json:"total,omitempty"`
	Page      int                        `json:"page,omitempty"`
}
