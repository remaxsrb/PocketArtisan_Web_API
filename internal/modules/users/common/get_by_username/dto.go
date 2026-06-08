package getbyusername

type GetCraftsmanByUsernameResponse struct {
	Username        string  `json:"username"`
	ProfilePicture  string  `json:"profilePicture"`
	Firstname       string  `json:"firstname"`
	Lastname        string  `json:"lastname"`
	Craft           string  `json:"craft,omitempty"`
	Rating          float64 `json:"rating,omitempty"`
	NumberOfRatings int     `json:"numberOfRatings,omitempty"`
	Email           string  `json:"email"`
	Gender          string  `json:"gender"`
}

type GetUserByUsernameResponse struct {
	Username       string `json:"username"`
	ProfilePicture string `json:"profilePicture"`
	Firstname      string `json:"firstname"`
	Lastname       string `json:"lastname"`
	Email          string `json:"email"`
	Gender         string `json:"gender"`
}
