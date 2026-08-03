package rate

import (
	prodmod "PocketArtisan/internal/modules/product"
	"PocketArtisan/internal/modules/utils"
	"context"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  prodmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: prodmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req Request) (Response, error) {
	customerID := ctx.Value("user_id").(uint64)

	product, err := uc.repo.RateProduct(ctx, uint64(req.ProductID), customerID, int(req.Rating))
	if err != nil {
		return Response{}, err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "products")

	return Response{
		AverageRating:   product.Rating,
		NumberOfRatings: product.NumberOfRatings,
	}, nil
}
