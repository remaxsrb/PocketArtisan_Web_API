package create

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

func (uc *UseCase) Execute(ctx context.Context, req CreateProductRequest) (*product.Product, error) {
	var existing product.Product

	if err := uc.db.WithContext(ctx).Where("name = ?", req.Name).First(&existing).Error; err == nil {
		return nil, errors.New("product already exists")
	}

	new_product := &product.Product{
		Name:  req.Name,
		Price: req.Price,
	}

	if err := uc.db.Create(new_product).Error; err != nil {
		return nil, err
	}

	return new_product, nil
}
