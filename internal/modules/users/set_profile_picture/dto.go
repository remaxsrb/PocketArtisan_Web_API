package set_profile_picture

type SetProfilePictureRequest struct {
	Username          string `json:"username" binding:"required"`
	NewProfilePicture string `json:"new_profile_picture" binding:"required"`
}
