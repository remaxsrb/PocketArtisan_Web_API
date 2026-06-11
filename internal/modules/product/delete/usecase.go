package delete

import (
	"PocketArtisan/internal/modules/product"
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

	var existing product.Product

	if err := uc.db.WithContext(ctx).Where("id = ?", req.ProductID).First(&existing).Error; err != nil {
		return errors.New("product not found")
	}
	

	if existing.CraftsmanID != req.CraftsmanID {
		return errors.New("forbidden: product does not belong to this craftsman")
	}

	if err := uc.db.WithContext(ctx).Where("product_id = ?", existing.ID).Delete(&product.ProductImage{}).Error; err != nil {
		return err
	}

	if err := uc.db.WithContext(ctx).Where("product_id = ?", existing.ID).Delete(&product.ProductVideo{}).Error; err != nil {
		return err
	}

	if err := uc.db.WithContext(ctx).Delete(&existing).Error; err != nil {
		return err
	}

	return nil
}