package users

type CraftsmanResponse struct {
	ID              int64   `json:"craftsman_id" omitempty:"id"`
	Firstname       string  `json:"firstname"`
	Lastname        string  `json:"lastname"`
	Username        string  `json:"username" gorm:"column:username"`
	Email           string  `json:"email"`
	ProfilePicture  string  `json:"profilePicture"`
	Gender          string  `json:"gender"`
	Craft           string  `json:"craft"`
	Rating          float64 `json:"rating"`
	NumberOfRatings int     `json:"numberOfRatings"`
	Role            string  `json:"role" omitempty:"role"`
}

type RegularUserResponse struct {
	ID             int64  `json:"id" omitempty:"id"`
	Firstname      string `json:"firstname"`
	Lastname       string `json:"lastname"`
	Username       string `json:"username" gorm:"column:username"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profilePicture"`
	Gender         string `json:"gender"`
	Role           string `json:"role" omitempty:"role"`
}
