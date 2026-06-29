package by_rating

import (
	"PocketArtisan/internal/modules/users"
)

type SortDtoRequest struct {
	Limit int `json:"limit"`
	Skip  int `json:"skip"`
}

type SortCraftsmenResponse struct {
	Craftsmen []*users.CraftsmanResponse `json:"craftsmen"`
	Total     int64                      `json:"total,omitempty"`
	Page      int                        `json:"page,omitempty"`
}
