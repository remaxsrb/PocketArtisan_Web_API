package users

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             uint64    `json:"id" gorm:"primaryKey"`
	Username       string    `json:"username" gorm:"unique;not null"`
	Email          string    `json:"email" gorm:"unique;not null"`
	Firstname      string    `json:"firstname"`
	Lastname       string    `json:"lastname"`
	DateOfBirth    time.Time `json:"date_of_birth" gorm:"type:date"`
	PasswordHash   string    `json:"-" gorm:"not null"`
	ProfilePicture string    `json:"profile_picture"`
	Gender         string    `json:"gender" gorm:"not null"`
	Role           string    `json:"role" gorm:"not null;default:'user'"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	LastLoginAt    time.Time `json:"last_login_at" gorm:"autoUpdateTime"`

	Craftsman *Craftsman `json:"craftsman" gorm:"foreignKey:UserID"`
}

type Craftsman struct {
	ID              uint64  `json:"id" gorm:"primaryKey"`
	UserID          uint64  `json:"user_id" gorm:"unique;not null"` // <--- The actual DB Foreign Key
	Craft           string  `json:"craft" gorm:"index;not null"`
	Rating          float64 `json:"rating" gorm:"not null;default:0.0"`
	NumberOfRatings int     `json:"number_of_ratings" gorm:"not null;default:0"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
