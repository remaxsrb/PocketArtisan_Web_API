package register

import "time"

type RegisterRequest struct {
	Username    string    `json:"username" binding:"required"`
	Email       string    `json:"email" binding:"required,email"`
	Password    string    `json:"password" binding:"required"`
	Firstname   string    `json:"firstname"`
	Lastname    string    `json:"lastname"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender" binding:"required"`
}

type RegisterResponse struct {
	Username string `json:"username"`
}
