package toggle_hide

import (
	prodmod "PocketArtisan/internal/modules/product"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  prodmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: prodmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req ToggleHideProduct) error {

	existing, err := uc.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return errors.New("product not found")
	}

	if existing.CraftsmanID != req.CraftsmanID {
		return errors.New("forbidden: product does not belong to this craftsman")
	}

	existing.Hidden = !existing.Hidden

	if err := uc.repo.Save(ctx, existing); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "products")

	return nil
}
