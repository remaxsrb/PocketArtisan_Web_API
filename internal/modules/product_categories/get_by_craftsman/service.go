package getbycraftsman

import (
	"PocketArtisan/internal/modules/product"
	pcmod "PocketArtisan/internal/modules/product_categories"
	"PocketArtisan/internal/modules/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"PocketArtisan/internal/entities"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo           pcmod.Repository
	cache          *redis.Client
	productService product.Service
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: pcmod.NewGormRepository(db), cache: cache, productService: product.NewService(db)}
}

func (uc *Service) Execute(ctx context.Context, username string) ([]entities.ProductCategory, error) {
	const cacheTTL = 5 * time.Minute

	craftsman, err := uc.productService.GetCraftsmanByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "product_categories")
	cacheKey := fmt.Sprintf("product_categories:craft:v:%d:%d", cacheVersion, craftsman.CraftID)

	if cached, err := uc.cache.Get(ctx, cacheKey).Result(); err == nil {
		var pcs []entities.ProductCategory
		if err := json.Unmarshal([]byte(cached), &pcs); err == nil {
			return pcs, nil
		}
	}

	pcs, err := uc.repo.FindByCraftID(ctx, craftsman.CraftID)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(pcs); err == nil {
		_ = uc.cache.Set(ctx, cacheKey, data, cacheTTL).Err()
	}

	return pcs, nil
}
