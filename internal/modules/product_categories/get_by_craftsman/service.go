package getbycraftsman

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/product"
	"PocketArtisan/internal/modules/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	db             *gorm.DB
	cache          *redis.Client
	productService product.Service
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{db: db, cache: cache, productService: product.NewService(db)}
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

	var pcs []entities.ProductCategory
	if err := uc.db.WithContext(ctx).
		Table("product_categories").
		Joins("INNER JOIN craft_product_categories ON craft_product_categories.category_id = product_categories.id").
		Where("craft_product_categories.craft_id = ?", craftsman.CraftID).
		Find(&pcs).Error; err != nil {
		return nil, err
	}

	if data, err := json.Marshal(pcs); err == nil {
		_ = uc.cache.Set(ctx, cacheKey, data, cacheTTL).Err()
	}

	return pcs, nil
}