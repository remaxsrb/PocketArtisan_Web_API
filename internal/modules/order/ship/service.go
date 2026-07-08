package ship

import (
	"context"
	"errors"
	"fmt"
	"time"

	"PocketArtisan/internal/entities"
	ordermod "PocketArtisan/internal/modules/order"
	"PocketArtisan/internal/modules/payment"
	"PocketArtisan/internal/modules/utils"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo    ordermod.Repository
	cache   *redis.Client
	gateway payment.Gateway
}

func NewService(db *gorm.DB, cache *redis.Client, gw payment.Gateway) *Service {
	return &Service{repo: ordermod.NewGormRepository(db), cache: cache, gateway: gw}
}

func (uc *Service) Execute(ctx context.Context, req ShipOrderRequest) (entities.OrderStatus, error) {

	existing, err := uc.repo.FindByID(ctx, req.OrderID)
	if err != nil {
		return "", errors.New("order not found")
	}

	if existing.CraftsmanID != req.CraftsmanID {
		return "", errors.New("forbidden: order does not belong to this craftsman")
	}

	if existing.CustomerID != req.CustomerID {
		return "", errors.New("forbidden: order does not belong to this customer")
	}

	if existing.PaymentType == entities.PaymentCreditCard && existing.PaymentReservationID != "" {
		if err := uc.gateway.Capture(ctx, existing.PaymentReservationID); err != nil {
			return "", fmt.Errorf("capture payment: %w", err)
		}
	}

	nextStatus, err := ordermod.NextOrderStatus(existing.Status, ordermod.OrderActionShip)
	if err != nil {
		return "", err
	}
	existing.Status = nextStatus
	now := time.Now()
	existing.ShippedAt = &now
	existing.CompletedAt = &now

	if err := uc.repo.Save(ctx, existing); err != nil {
		return "", err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "orders")

	return nextStatus, nil
}
