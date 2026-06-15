package set_role

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{db: db, cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req SetRoleRequest) error {
	var user entities.User

	var isAllowedRole bool = req.Role == "craftsman" || req.Role == "user" || req.Role == "admin"

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
