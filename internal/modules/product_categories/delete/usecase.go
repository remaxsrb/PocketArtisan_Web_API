package delete

import (
	"PocketArtisan/internal/modules/utils"
	"errors"
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"PocketArtisan/internal/modules/product_categories"
)

type UseCase struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewUseCase(db *gorm.DB, cache *redis.Client) *UseCase {
	return &UseCase{db: db, cache: cache}
}

func (uc *UseCase) Execute(ctx context.Context, req DeleteProductCategoryRequest) error {

	var pc product_categories.ProductCategory
	if err := uc.db.WithContext(ctx).Where("name = ?", req.Name).First(&pc).Error; err != nil {
		return errors.New("product category does not exist")
	}

	if err := uc.db.WithContext(ctx).Delete(&pc).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "products", "product_categories")

	return nil
}