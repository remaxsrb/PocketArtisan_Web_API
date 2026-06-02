package craftsman_application

import "time"

type CraftsmanApplication struct {
	ID         uint64    `json:"id" gorm:"primary_key"`
	UserID     uint64    `json:"user_id"`
	Status     string    `json:"status" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	ResolvedAt time.Time `json:"resolved_at"`
}

const (
	StatusPending  string = "pending"
	StatusRejected string = "rejected"
	StatusAccepted string = "accepted"
)
