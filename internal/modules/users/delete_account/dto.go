package delete_account

type DeleteAccountRequest struct {
	Username string `json:"username" binding:"required"`
}
