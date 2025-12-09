package common

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Username       string    `json:"username" gorm:"unique;not null"`
	Email          string    `json:"email" gorm:"unique;not null"`
	Firstname      string    `json:"first_name"`
	Lastname       string    `json:"last_name"`
	DateOfBirth    time.Time `json:"date_of_birth" gorm:"type:date"`
	PasswordHash   string    `json:"-" gorm:"not null"`
	ProfilePicture string    `json:"profile_picture"`
	Gender         string    `json:"gender" gorm:"not null"`
	Role           string    `json:"role" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
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
