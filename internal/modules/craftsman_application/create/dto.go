package create

type CraftsmanApplicationRequest struct {
	Email     string `json:"email"`
	Craft     string `json:"craft"`
	ResumeURL string `json:"resume_url" binding:"required"`
}
