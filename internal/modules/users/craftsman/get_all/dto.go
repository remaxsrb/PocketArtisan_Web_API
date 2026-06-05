package getall

import (
	"PocketArtisan/internal/modules/users"
)

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
	Craftsmen []users.User `json:"craftsmen"`
	Total     int64        `json:"total,omitempty"`
	Page      int          `json:"page,omitempty"` // Current page number (derived from skip/limit)

}
