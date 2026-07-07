package change_password

import (
	usersmod "PocketArtisan/internal/modules/users"
	"PocketArtisan/internal/modules/utils"
	"PocketArtisan/internal/validators"
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

func (uc *Service) Execute(ctx context.Context, req ChangePasswordRequest) error {

	if err := validators.ValidatePassword(req.NewPassword); err != nil {
		return errors.New(err.Error())
	}

	user, err := uc.repo.FindUserByID(ctx, req.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		return err
	}

	if err := uc.repo.SaveUser(ctx, user); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users")

	return nil
}
