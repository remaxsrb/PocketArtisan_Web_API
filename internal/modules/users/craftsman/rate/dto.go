package rate

type Request struct {
	Username string `json:"username" binding:"required"`
	Rating   int8   `json:"rating" binding:"required"`
}
