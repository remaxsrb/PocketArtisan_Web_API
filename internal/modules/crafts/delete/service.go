package delete

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

func (uc *Service) Execute(ctx context.Context, req DeleteCraftRequest) error {

	var c entities.Craft
	if err := uc.db.WithContext(ctx).Where("name = ?", req.Name).First(&c).Error; err != nil {
		return errors.New("craft does not exist")
	}

	if err := uc.db.WithContext(ctx).Delete(&c).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "craftsmen", "crafts")

	return nil

}
