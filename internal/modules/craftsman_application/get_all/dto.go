package get_all

import (
	"PocketArtisan/internal/entities"
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
	CraftsmanApplications []*entities.CraftsmanApplication `json:"craftsman_applications"`
	Total                 int64                            `json:"total,omitempty"`
	Page                  int                              `json:"page,omitempty"`
}
