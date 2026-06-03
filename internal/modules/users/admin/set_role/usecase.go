package set_role

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

func (uc *UseCase) Execute(ctx context.Context, req SetRoleRequest) error {
	var user users.User

	if req.Role != "artisan" {
		return errors.New("only role which admin can set is artisan")
	}

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&user).Error; err != nil {
		return errors.New("users not found")
	}

	user.Role = req.Role

	if err := uc.db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}
