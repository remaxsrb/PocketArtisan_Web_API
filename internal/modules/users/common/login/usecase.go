package login

import (
	"PocketArtisan/internal/modules/users"
	"context"
	"errors"
	"strconv"
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

func (uc *UseCase) Execute(ctx context.Context, req LoginRequest) (*LoginResponse, error) {

	var existing users.User

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&existing).Error; err != nil {
		return nil, errors.New("username not found")
	}

	if !existing.CheckPassword(req.Password) {
		return nil, errors.New("invalid password")
	}

	existing.LastLoginAt = time.Now()
	uc.db.WithContext(ctx).Save(&existing)

	r := LoginResponse{ID: strconv.FormatUint(existing.ID, 10), Username: existing.Username, Role: existing.Role,
		Firstname: existing.Firstname, Lastname: existing.Lastname, ProfilePicture: existing.ProfilePicture,
		Craft: existing.Craft, Rating: existing.Rating, NumberOfRatings: existing.NumberOfRatings, Email: existing.Email}

	return &r, nil

}
