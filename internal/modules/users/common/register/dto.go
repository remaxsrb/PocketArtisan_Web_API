package register

type RegisterRequest struct {
	Username       string `json:"username" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required"`
	Firstname      string `json:"firstname"`
	Lastname       string `json:"lastname"`
	DateOfBirth    string `json:"date_of_birth"`
	Gender         string `json:"gender" binding:"required"`
	Role           string `json:"role,omitempty"`
	TurnstileToken string `json:"turnstile_token"`
}

type RegisterResponse struct {
	Username string `json:"username"`
}
