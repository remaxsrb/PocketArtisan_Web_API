package entities

import "time"

type CraftsmanApplication struct {
	ID         uint64     `json:"id" gorm:"primary_key"`
	Email      string     `json:"email" gorm:"not null;index"`
	Status     string     `json:"status" gorm:"not null"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	Craft      string     `json:"craft" gorm:"not null"`
	ResumeURL  string     `json:"resume_url" gorm:"not null"`
}

const (
	StatusPending  string = "pending"
	StatusRejected string = "rejected"
	StatusAccepted string = "approved"
)
