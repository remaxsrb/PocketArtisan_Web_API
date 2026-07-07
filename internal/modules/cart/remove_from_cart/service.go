package removefromcart

import (
	cartmod "PocketArtisan/internal/modules/cart"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  cartmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: cartmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req RemoveFromCartRequest) (*cartmod.CartResponse, error) {

	userCart, err := uc.repo.GetUserCart(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := uc.repo.DeleteCartItem(ctx, userCart.ID, req.ProductID)
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, errors.New("item not found in cart")
	}

	if err := uc.repo.RefreshCartTotal(ctx, userCart.ID); err != nil {
		return nil, err
	}

	updatedCart, err := uc.repo.GetUserCart(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &cartmod.CartResponse{Cart: *updatedCart}, nil
}
