package create


type Request struct {
	Email string `json:"email"`
	Craft  string `json:"craft"`
}

type Response struct {
	Message string `json:"message"`
}