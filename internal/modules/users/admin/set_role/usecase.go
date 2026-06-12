package set_role

import (
	"PocketArtisan/internal/modules/utils"
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

	var isAllowedRole bool= req.Role == "craftsman" || req.Role == "user" || req.Role == "admin"

	if !isAllowedRole {
		return errors.New("only roles which admin can set are craftsman, user, or admin")
	}

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&user).Error; err != nil {
		return errors.New("users not found")
	}

	user.Role = req.Role

	if err := uc.db.Save(&user).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users", "craftsmen")

	return nil
}
