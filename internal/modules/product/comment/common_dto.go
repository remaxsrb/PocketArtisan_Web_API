package comment

import "time"

type Response struct {
	Username   string    `json:"username"`
	ProductID  uint64    `json:"productId"`
	Text       string    `json:"text"`
	PhotoLinks []string  `json:"photoLinks"`
	VideoLinks []string  `json:"videoLinks"`
	CreatedAt  time.Time `json:"createdAt"`
}

func ToResponse(c *Comment) *Response {
	return &Response{
		Username:   c.Username,
		ProductID:  c.ProductID,
		Text:       c.Text,
		PhotoLinks: c.PhotoLinks,
		VideoLinks: c.VideoLinks,
		CreatedAt:  c.CreatedAt,
	}
}
