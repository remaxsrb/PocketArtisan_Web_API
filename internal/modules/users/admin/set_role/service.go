package set_role

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

func (uc *Service) Execute(ctx context.Context, req SetRoleRequest) error {

	var isAllowedRole bool = req.Role == "craftsman" || req.Role == "user" || req.Role == "admin"

	if !isAllowedRole {
		return errors.New("only roles which admin can set are craftsman, user, or admin")
	}

	user, err := uc.repo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		return errors.New("users not found")
	}

	user.Role = req.Role

	if err := uc.repo.SaveUser(ctx, user); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users", "craftsmen")

	return nil
}
