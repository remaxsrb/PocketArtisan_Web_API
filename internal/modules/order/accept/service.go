package accept

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

func (uc *Service) Execute(ctx context.Context, req AcceptOrderRequest) (entities.OrderStatus, error) {

	var existing entities.Order

	if err := uc.db.WithContext(ctx).Where("id = ?", req.OrderID).First(&existing).Error; err != nil {
		return "", errors.New("order not found")
	}

	if existing.CraftsmanID != req.CraftsmanID {
		return "", errors.New("forbidden: order does not belong to this craftsman")
	}

	existing.Status = entities.OrderAccepted

	if err := uc.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return "", err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "orders")

	return existing.Status, nil
}