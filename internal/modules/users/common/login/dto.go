package login

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResult struct {
	ID       uint64
	Role     string
	Response interface{}
}
