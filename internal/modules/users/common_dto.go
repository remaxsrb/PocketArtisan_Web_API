package users

/*

 Go does not support inheritance but it does support struct embedding like C.
Therefore refactoring reponses to use embedded structs make sense for cleaner and more readable code

*/

type UserResponse struct {
	Firstname      string `json:"firstname"`
	Lastname       string `json:"lastname"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profilePicture"`
	Gender         string `json:"gender"`
}

type RegularUserResponse struct {
	UserResponse
}

type CraftsmanResponse struct {
	UserResponse
	CraftsmanID      uint64  `json:"craftsmanId"`
	Craft           string  `json:"craft"`
	Rating          float64 `json:"rating"`
	NumberOfRatings int     `json:"numberOfRatings"`
}
