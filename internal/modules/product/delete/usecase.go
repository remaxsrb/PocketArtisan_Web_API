package delete

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

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

func (uc *UseCase) Execute(ctx context.Context, req DeleteProductRequest) error {

	var existing entities.Product

	if err := uc.db.WithContext(ctx).Where("id = ?", req.ProductID).First(&existing).Error; err != nil {
		return errors.New("product not found")
	}

	if existing.CraftsmanID != req.CraftsmanID {
		return errors.New("forbidden: product does not belong to this craftsman")
	}

	if err := uc.db.WithContext(ctx).Where("product_id = ?", existing.ID).Delete(&entities.ProductImage{}).Error; err != nil {
		return err
	}

	if err := uc.db.WithContext(ctx).Where("product_id = ?", existing.ID).Delete(&entities.ProductVideo{}).Error; err != nil {
		return err
	}

	if err := uc.db.WithContext(ctx).Delete(&existing).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "products")

	return nil
}
