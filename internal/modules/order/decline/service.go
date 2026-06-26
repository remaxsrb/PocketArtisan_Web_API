package decline

import (
	"context"
	"errors"
	"fmt"

	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/payment"
	"PocketArtisan/internal/modules/utils"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	db      *gorm.DB
	cache   *redis.Client
	gateway payment.Gateway
}

func NewService(db *gorm.DB, cache *redis.Client, gw payment.Gateway) *Service {
	return &Service{db: db, cache: cache, gateway: gw}
}

func (uc *Service) Execute(ctx context.Context, req DeclineOrderRequest) (entities.OrderStatus, error) {

	var existing entities.Order

	if err := uc.db.WithContext(ctx).Where("id = ?", req.OrderID).First(&existing).Error; err != nil {
		return "", errors.New("order not found")
	}

	if existing.CraftsmanID != req.CraftsmanID {
		return "", errors.New("forbidden: order does not belong to this craftsman")
	}

	if existing.PaymentType == entities.PaymentCreditCard && existing.PaymentReservationID != "" {
		if err := uc.gateway.Refund(ctx, existing.PaymentReservationID); err != nil {
			return "", fmt.Errorf("refund payment: %w", err)
		}
	}

	existing.Status = entities.OrderDeclined

	if err := uc.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return "", err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "orders")

	return existing.Status, nil
}
