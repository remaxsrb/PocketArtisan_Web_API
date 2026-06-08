package getall

import "PocketArtisan/internal/modules/users"

type Direction string

const (
	Next Direction = "next"
	Prev Direction = "prev"
)

type GetAllRequest struct {
	Limit int `form:"limit" query:"limit"`
	Skip  int `form:"skip" query:"skip"`
}

type GetAllResponse struct {
	Craftsmen []*users.CraftsmanResponse `json:"craftsmen"`
	Total     int64                      `json:"total,omitempty"`
	Page      int                        `json:"page,omitempty"`
}
