package delete

import (
	pcmod "PocketArtisan/internal/modules/product_categories"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  pcmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: pcmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req DeleteProductCategoryRequest) error {

	pc, err := uc.repo.FindByName(ctx, req.Name)
	if err != nil {
		return errors.New("product category does not exist")
	}

	if err := uc.repo.Delete(ctx, pc); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "products", "product_categories")

	return nil
}
