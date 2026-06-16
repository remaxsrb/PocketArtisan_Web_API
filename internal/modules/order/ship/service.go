package ship

import (
	"context"
	"errors"
		
	"PocketArtisan/internal/entities"
	"gorm.io/gorm"
	"github.com/go-redis/redis/v8"

	"PocketArtisan/internal/modules/utils"
)

type Service struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{db: db, cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req ShipOrderRequest) (entities.OrderStatus, error) {

	var existing entities.Order

	if err := uc.db.WithContext(ctx).Where("id = ?", req.OrderID).First(&existing).Error; err != nil {
		return "", errors.New("order not found")
	}

	if existing.CraftsmanID != req.CraftsmanID {
		return "", errors.New("forbidden: order does not belong to this craftsman")
	}

	if existing.CustomerID != req.CustomerID {
		return "", errors.New("forbidden: order does not belong to this customer")
	}

	existing.Status = entities.OrderShipped

	if err := uc.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return "", err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "orders")

	return existing.Status, nil
}
