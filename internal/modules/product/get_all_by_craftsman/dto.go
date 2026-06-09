package getallbycraftsman

import "PocketArtisan/internal/modules/product"

type Direction string

const (
	Next Direction = "next"
	Prev Direction = "prev"
)

type GetAllRequest struct {
	CraftsmanID uint64 `form:"craftsmanId" query:"craftsmanId"`
	Limit int `form:"limit" query:"limit"`
	Skip  int `form:"skip" query:"skip"`
}

type GetAllResponse struct {
	Products []*product.ProductResponse `json:"products"`
	Total    int64                      `json:"total,omitempty"`
	Page     int                        `json:"page,omitempty"`
}