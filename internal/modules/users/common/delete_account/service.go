package delete_account

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

func (uc *Service) Execute(ctx context.Context, req DeleteAccountRequest) error {

	var existing entities.User

	if err := uc.db.WithContext(ctx).Where("id = ?", req.UserID).First(&existing).Error; err != nil {
		return errors.New("user not found")
	}

	if err := uc.db.Delete(&existing).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users", "craftsmen", "products")

	return nil
}
