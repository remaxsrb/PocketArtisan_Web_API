package set_profile_picture

type SetProfilePictureRequest struct {
	NewProfilePicture string `json:"newProfilePicture" binding:"required"`
	UserID            uint64 `json:"-"`
}
