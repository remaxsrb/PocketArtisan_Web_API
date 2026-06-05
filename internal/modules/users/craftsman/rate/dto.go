package rate

type Request struct {
	UserID int `json:"user_id" binding:"required"`
	Rating   int8   `json:"rating" binding:"required"`
}
