package list

import "PocketArtisan/internal/modules/product/comment"

type Request struct {
	ProductID uint64
	Skip      int64
	Limit     int64
}

type Response struct {
	Comments []*comment.Response `json:"comments"`
	Total    int64               `json:"total"`
}
