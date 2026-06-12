package getbycategory

import "PocketArtisan/internal/modules/product"

type GetByCategoryRequest struct {
	Search string `form:"search" query:"search" required:"true"`
	Limit  int    `form:"limit" query:"limit"`
	Skip   int    `form:"skip" query:"skip"`
}

type GetByCategoryResponse struct {
	Products []*product.ProductResponse `json:"products"`
	Total    int64                      `json:"total,omitempty"`
	Page     int                        `json:"page,omitempty"`
}
