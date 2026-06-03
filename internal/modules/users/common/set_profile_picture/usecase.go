package set_profile_picture

import (
	"PocketArtisan/internal/modules/users"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UseCase struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewUseCase(db *gorm.DB, cache *redis.Client) *UseCase {
	return &UseCase{db: db, cache: cache}
}

func (uc *UseCase) Execute(ctx context.Context, req SetProfilePictureRequest) error {

	var existing users.User

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&existing).Error; err != nil {
		return errors.New("username not found")
	}

	existing.ProfilePicture = req.NewProfilePicture

	if err := uc.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return err
	}

	return nil

}
