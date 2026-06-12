package get_all

import (
	"PocketArtisan/internal/modules/product_categories"
	"context"

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

	var pcs []product_categories.ProductCategory
	if err := uc.db.WithContext(ctx).Find(&pcs).Error; err != nil {
		return nil, err
	}

	return pcs, nil

}