package get_all

import (
	"PocketArtisan/internal/modules/product_categories"
	"PocketArtisan/internal/modules/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

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

func (uc *UseCase) Execute(ctx context.Context) ([]product_categories.ProductCategory, error) {
	const cacheTTL = 5 * time.Minute

	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "product_categories")
	cacheKey := fmt.Sprintf("product_categories:all:v:%d", cacheVersion)

	if cached, err := uc.cache.Get(ctx, cacheKey).Result(); err == nil {
		var pcs []product_categories.ProductCategory
		if err := json.Unmarshal([]byte(cached), &pcs); err == nil {
			return pcs, nil
		}
	}

	var pcs []product_categories.ProductCategory
	if err := uc.db.WithContext(ctx).Find(&pcs).Error; err != nil {
		return nil, err
	}

	if data, err := json.Marshal(pcs); err == nil {
		_ = uc.cache.Set(ctx, cacheKey, data, cacheTTL).Err()
	}

	return pcs, nil

}