package delete_account

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

func (uc *Service) Execute(ctx context.Context, req DeleteAccountRequest) error {

	existing, err := uc.repo.FindUserByID(ctx, req.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := uc.repo.DeleteUser(ctx, existing); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users", "craftsmen", "products")

	return nil
}
