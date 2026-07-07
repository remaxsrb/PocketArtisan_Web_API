package rate

type Request struct {
	CraftsmanID int  `json:"craftsmanID" binding:"required"`
	Rating      int8 `json:"rating" binding:"required"`
}

type Response struct {
	AverageRating   float64 `json:"averageRating"`
	NumberOfRatings int     `json:"numberOfRatings"`
}
