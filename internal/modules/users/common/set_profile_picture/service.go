package set_profile_picture

import (
	usersmod "PocketArtisan/internal/modules/users"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  usersmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: usersmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req SetProfilePictureRequest) error {

	existing, err := uc.repo.FindUserByID(ctx, req.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	existing.ProfilePicture = req.NewProfilePicture

	if err := uc.repo.SaveUser(ctx, existing); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users")

	return nil

}
