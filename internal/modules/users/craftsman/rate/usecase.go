package rate

import (
	"PocketArtisan/internal/modules/users"
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

func (uc *UseCase) Execute(ctx context.Context, req Request) (Response, error) {
	var craftsman users.Craftsman

	if err := uc.db.WithContext(ctx).Where("user_id = ?", req.UserID).First(&craftsman).Error; err != nil {
		return Response{}, errors.New("craftsman not found")
	}

	new_avg_rating := ((craftsman.Rating * float64(craftsman.NumberOfRatings)) + float64(req.Rating)) / float64(craftsman.NumberOfRatings+1)

	craftsman.Rating = new_avg_rating
	craftsman.NumberOfRatings++

	if err := uc.db.Save(&craftsman).Error; err != nil {
		return Response{}, err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users", "craftsmen")

	return Response{
		AverageRating:   craftsman.Rating,
		NumberOfRatings: craftsman.NumberOfRatings,
	}, nil
}
