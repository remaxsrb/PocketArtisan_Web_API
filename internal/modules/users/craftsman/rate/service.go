package rate

import (
	usersmod "PocketArtisan/internal/modules/users"
	"PocketArtisan/internal/modules/utils"
	"context"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  usersmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: usersmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req Request) (Response, error) {
	customerID := ctx.Value("user_id").(uint64)

	craftsman, err := uc.repo.RateCraftsman(ctx, uint64(req.CraftsmanID), customerID, int(req.Rating))
	if err != nil {
		return Response{}, err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users", "craftsmen")

	return Response{
		AverageRating:   craftsman.Rating,
		NumberOfRatings: craftsman.NumberOfRatings,
	}, nil
}
