package login

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	ID              string  `json:"id"`
	Role            string  `json:"role"`
	Username        string  `json:"username"`
	Firstname       string  `json:"first_name"`
	Lastname        string  `json:"last_name"`
	ProfilePicture  string  `json:"profile_picture"`
	Craft           string  `json:"craft"`
	Rating          float64 `json:"rating"`
	NumberOfRatings int     `json:"number_of_ratings"`
	Email           string  `json:"email"`
}
