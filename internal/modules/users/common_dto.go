package users

import (
	"PocketArtisan/internal/entities"
)

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
	Role           string `json:"role"`
	CreatedAt      string `json:"created_at"`
}

type RegularUserResponse struct {
	UserResponse
	Cart *entities.Cart `json:"cart,omitempty"`
}

type CraftsmanResponse struct {
	UserResponse
	CraftsmanID     uint64  `json:"craftsmanId"`
	Craft           string  `json:"craft"`
	Rating          float64 `json:"rating"`
	NumberOfRatings int     `json:"numberOfRatings" gorm:"column:number_of_ratings"`
	Biography       string  `json:"biography"`
}
