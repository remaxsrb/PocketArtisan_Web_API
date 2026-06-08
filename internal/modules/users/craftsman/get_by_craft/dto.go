package getbycraft

type GetByCraftRequest struct {
	Limit int
	Skip  int
}

type CraftsmanResponse struct {
	Firstname       string  `json:"firstname"`
	Lastname        string  `json:"lastname"`
	Username        string  `json:"username" gorm:"column:username"`
	Email           string  `json:"email"`
	ProfilePicture  string  `json:"profilePicture"`
	Craft           string  `json:"craft"`
	Rating          float64 `json:"rating"`
	NumberOfRatings int     `json:"numberRatings"`
}

type GetByCraftResponse struct {
	Craftsmen []*CraftsmanResponse `json:"craftsmen"`
	Total     int64                `json:"total,omitempty"`
	Page      int                  `json:"page,omitempty"`
}