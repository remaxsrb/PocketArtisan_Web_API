package approve

type Request struct {
	ApplicationID int    `json:"id"`
	Message       string `json:"message"`
}
