package toggle_hide

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

func (uc *Service) Execute(ctx context.Context, req ToggleHideProduct) error {

	var existing entities.Product

	if err := uc.db.WithContext(ctx).Where("id = ?", req.ProductID).First(&existing).Error; err != nil {
		return errors.New("product not found")
	}

	if existing.CraftsmanID != req.CraftsmanID {
		return errors.New("forbidden: product does not belong to this craftsman")
	}

	existing.Hidden = !existing.Hidden

	uc.db.WithContext(ctx).Save(&existing)

	utils.BumpCacheVersion(ctx, uc.cache, "products")

	return nil
}
