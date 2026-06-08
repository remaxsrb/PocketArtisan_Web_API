package login

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	ID              string  `json:"id"`
	Role            string  `json:"role"`
	Username        string  `json:"username"`
	Firstname       string  `json:"firstname"`
	Lastname        string  `json:"lastname"`
	ProfilePicture  string  `json:"profilePicture"`
	Craft           string  `json:"craft"`
	Rating          float64 `json:"rating"`
	Email           string  `json:"email"`
}
