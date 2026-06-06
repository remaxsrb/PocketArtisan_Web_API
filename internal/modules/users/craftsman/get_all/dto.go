package getall

type Direction string

const (
	Next Direction = "next"
	Prev Direction = "prev"
)

type GetAllRequest struct {
	Limit int `form:"limit" query:"limit"`
	Skip  int `form:"skip" query:"skip"`
}

type CraftsmanResponse struct {
	Firstname       string  `json:"firstname"`
	Lastname        string  `json:"lastname"`
	Email           string  `json:"email"`
	ProfilePicture  string  `json:"profile_picture"`
	Craft           string  `json:"craft"`
	Rating          float64 `json:"rating"`
	NumberOfRatings int     `json:"number_of_ratings"`
}

type GetAllResponse struct {
	Craftsmen []*CraftsmanResponse `json:"craftsmen"`
	Total     int64                `json:"total,omitempty"`
	Page      int                  `json:"page,omitempty"`
}
