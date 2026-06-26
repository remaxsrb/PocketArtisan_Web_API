package rate

import (
	"context"
	"errors"

	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{db: db, cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req Request) (Response, error) {
	customerID := ctx.Value("user_id").(uint64)

	var craftsman entities.Craftsman
	var resp Response

	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", req.UserID).First(&craftsman).Error; err != nil {
			return errors.New("craftsman not found")
		}

		var existing entities.CraftsmanRatingRecord
		err := tx.Where("customer_id = ? AND craftsman_id = ?", customerID, uint64(req.UserID)).First(&existing).Error
		if err == nil {
			return errors.New("you have already rated this craftsman")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		record := entities.CraftsmanRatingRecord{
			CustomerID:  customerID,
			CraftsmanID: uint64(req.UserID),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		craftsman.Rating = ((craftsman.Rating * float64(craftsman.NumberOfRatings)) + float64(req.Rating)) / float64(craftsman.NumberOfRatings+1)
		craftsman.NumberOfRatings++

		return tx.Save(&craftsman).Error
	})

	if err != nil {
		return Response{}, err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users", "craftsmen")

	resp.AverageRating = craftsman.Rating
	resp.NumberOfRatings = craftsman.NumberOfRatings
	return resp, nil
}
