package users

type CraftsmanResponse struct {
	Firstname       string  `json:"firstname"`
	Lastname        string  `json:"lastname"`
	Username        string  `json:"username" gorm:"column:username"`
	Email           string  `json:"email"`
	ProfilePicture  string  `json:"profilePicture"`
	Gender          string  `json:"gender"`
	Craft           string  `json:"craft"`
	Rating          float64 `json:"rating"`
	NumberOfRatings int     `json:"numberRatings"`
}

type RegularUserResponse struct {
	Firstname      string `json:"firstname"`
	Lastname       string `json:"lastname"`
	Username       string `json:"username" gorm:"column:username"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profilePicture"`
	Gender         string `json:"gender"`
}
