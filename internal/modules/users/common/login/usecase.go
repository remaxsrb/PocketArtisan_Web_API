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

func (uc *UseCase) Execute(ctx context.Context, req LoginRequest) (any, error) {

	var existing users.User

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&existing).Error; err != nil {
		return nil, errors.New("username not found")
	}

	if !existing.CheckPassword(req.Password) {
		return nil, errors.New("invalid password")
	}

	existing.LastLoginAt = time.Now()
	uc.db.WithContext(ctx).Save(&existing)

	isCraftsman := existing.Role == "craftsman"

	if isCraftsman {
		var r *users.CraftsmanResponse
		uc.db.WithContext(ctx).
			Table("users").
			Select(`
				users.username,
				users.firstname,
				users.lastname,
				users.email,
				users.profile_picture,
				users.gender,
				users.role,
				craftsmen.craft,
				craftsmen.rating,
				craftsmen.number_of_ratings
			`).
			Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
			Where("users.username = ?", existing.Username).
			Scan(&r)

		return r, nil
	}

	r := users.RegularUserResponse{Username: existing.Username, Role: existing.Role,
		Firstname: existing.Firstname, Lastname: existing.Lastname, ProfilePicture: existing.ProfilePicture,
		Email: existing.Email}

	return &r, nil

}
