package get_by_craftsman

import "PocketArtisan/internal/modules/order"

type Direction string

const (
	Next Direction = "next"
	Prev Direction = "prev"
)

type GetAllRequest struct {
	CraftsmanID uint64
	Limit       int `form:"limit" query:"limit"`
	Skip        int `form:"skip" query:"skip"`
}

type GetAllResponse struct {
	Orders []*order.OrderResponse `json:"orders"`
	Total  int64                  `json:"total,omitempty"`
	Page   int                    `json:"page,omitempty"`
}
