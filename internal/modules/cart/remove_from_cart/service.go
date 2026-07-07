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

	rowsAffected, err := uc.repo.DeleteCartItem(ctx, req.CartID, req.ProductID)
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, errors.New("item not found in cart")
	}

	price, err := uc.repo.GetProductPrice(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	userCart, err := uc.repo.GetUserCart(ctx, req.CartID)
	if err != nil {
		return nil, err
	}

	userCart.Total -= price * float64(req.Quantity)
	if err := uc.repo.SaveCart(ctx, userCart); err != nil {
		return nil, err
	}

	return &cartmod.CartResponse{Cart: *userCart}, nil
}
