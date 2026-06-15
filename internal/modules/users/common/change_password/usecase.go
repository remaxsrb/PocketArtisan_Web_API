package change_password

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
	"PocketArtisan/internal/validators"
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

func (uc *UseCase) Execute(ctx context.Context, req ChangePasswordRequest) error {
	var user entities.User

	if err := validators.ValidatePassword(req.NewPassword); err != nil {
		return errors.New(err.Error())
	}

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		return err
	}

	if err := uc.db.Save(&user).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users")

	return nil
}
