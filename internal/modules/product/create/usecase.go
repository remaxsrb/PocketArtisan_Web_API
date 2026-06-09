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

func (uc *UseCase) Execute(ctx context.Context, req NewProductRequest) (*product.ProductResponse, error) {
	var existing product.Product

	if err := uc.db.WithContext(ctx).Where("name = ? AND craftsman_id = ?", req.Name, req.CraftsmanID).First(&existing).Error; err == nil {
		return nil, errors.New("product already exists")
	}

	new_product := &product.Product{
		Name:          req.Name,
		CraftsmanID:   req.CraftsmanID,
		MaterialPrice: req.MaterialPrice,
		LaborPrice:    req.LaborPrice,
		Picture:       req.Picture,
		Hidden:        false,
		Description:   req.Description,
	}

	if err := uc.db.Create(new_product).Error; err != nil {
		return nil, err
	}

	response := &product.ProductResponse{
		Name:        new_product.Name,
		CraftsmanID: new_product.CraftsmanID,
		Hidden:      new_product.Hidden,
		Picture:     new_product.Picture,
		TotalPrice:  new_product.MaterialPrice + new_product.LaborPrice,
		Description: new_product.Description,
	}

	return response, nil
}
