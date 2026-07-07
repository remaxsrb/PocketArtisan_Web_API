package accept

import (
	"context"
	"errors"

	"PocketArtisan/internal/entities"
	ordermod "PocketArtisan/internal/modules/order"
	"PocketArtisan/internal/modules/utils"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  ordermod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: ordermod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req AcceptOrderRequest) (entities.OrderStatus, error) {

	existing, err := uc.repo.FindByID(ctx, req.OrderID)
	if err != nil {
		return "", errors.New("order not found")
	}

	if existing.CraftsmanID != req.CraftsmanID {
		return "", errors.New("forbidden: order does not belong to this craftsman")
	}

	nextStatus, err := ordermod.NextOrderStatus(existing.Status, ordermod.OrderActionAccept)
	if err != nil {
		return "", err
	}
	existing.Status = nextStatus

	if err := uc.repo.Save(ctx, existing); err != nil {
		return "", err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "orders")

	return nextStatus, nil
}
