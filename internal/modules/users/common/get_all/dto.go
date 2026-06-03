package get_all

import (
	"PocketArtisan/internal/modules/users"
	"time"
)

type Direction string

const (
	Next Direction = "next"
	Prev Direction = "prev"
)

type GetAllRequest struct {
	Limit     int        `json:"limit"`
	CursorAt  *time.Time `json:"cursor"`    // created_at
	ID        *uint64    `json:"id"`        // tiebreaker
	Direction Direction  `json:"direction"` // next | prev
}

type GetAllResponse struct {
	Users  []*users.User `json:"users"`
	NextAt *time.Time    `json:"next_at,omitempty"`
	NextID *uint64       `json:"next_id,omitempty"`
	PrevAt *time.Time    `json:"prev_at,omitempty"`
	PrevID *uint64       `json:"prev_id,omitempty"`
}
