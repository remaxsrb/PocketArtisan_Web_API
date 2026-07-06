package change_password

type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
	UserID      uint64 `json:"-"`
}
