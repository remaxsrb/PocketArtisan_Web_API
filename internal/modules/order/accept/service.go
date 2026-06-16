package accept

import (
	"context"
	"errors"
	"strconv"

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

func (uc *Service) Execute(ctx context.Context, req AcceptOrderRequest) error {

	var existing entities.Order

	if err := uc.db.WithContext(ctx).Where("id = ?", req.OrderID).First(&existing).Error; err != nil {
		return errors.New("order not found")
	}

	craftsmanID, err := strconv.ParseUint(req.CraftsmanID, 10, 64)
	if err != nil {
		return errors.New("invalid craftsman ID format")
	}

	if existing.CraftsmanID != craftsmanID {
		return errors.New("forbidden: order does not belong to this craftsman")
	}

	existing.Status = entities.OrderAccepted

	if err := uc.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "orders")

	return nil
}