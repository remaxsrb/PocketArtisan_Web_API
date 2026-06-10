package login

import (
	"PocketArtisan/internal/modules/users"
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UseCase struct {
	db     *gorm.DB
	cache  *redis.Client
	secret string
}

func NewUseCase(db *gorm.DB, cache *redis.Client) *UseCase {
	return &UseCase{db: db, cache: cache}
}

func (uc *UseCase) Execute(ctx context.Context, req LoginRequest) (LoginResult, error) {

	var existing users.User

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&existing).Error; err != nil {
		return LoginResult{}, errors.New("username not found")
	}

	if !existing.CheckPassword(req.Password) {
		return LoginResult{}, errors.New("invalid password")
	}

	existing.LastLoginAt = time.Now()
	uc.db.WithContext(ctx).Save(&existing)

	if existing.Role == "craftsman" {
		var r users.CraftsmanResponse
		uc.db.WithContext(ctx).
			Table("users").
			Select(`
                users.username,
                users.firstname,
                users.lastname,
                users.email,
                users.profile_picture,
                users.gender,
                craftsmen.craft,
                craftsmen.rating,
                craftsmen.number_of_ratings
            `).
			Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
			Where("users.username = ?", existing.Username).
			Scan(&r)
		return LoginResult{ID: existing.ID, Role: existing.Role, Response: &r}, nil
	}

	r := &users.RegularUserResponse{
		UserResponse: users.UserResponse{
			Username:       existing.Username,
			Firstname:      existing.Firstname,
			Lastname:       existing.Lastname,
			Email:          existing.Email,
			ProfilePicture: existing.ProfilePicture,
			Gender:         existing.Gender,
		},
	}
	return LoginResult{ID: existing.ID, Role: existing.Role, Response: r}, nil

}
