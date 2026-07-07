package addtocart

import (
	"PocketArtisan/internal/entities"
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

func (uc *Service) Execute(ctx context.Context, req AddToCartRequest) (*cartmod.CartResponse, error) {
	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	userCart, err := uc.repo.GetUserCart(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	existingItem, err := uc.repo.FindCartItem(ctx, userCart.ID, req.ProductID)
	if err == nil {
		existingItem.Quantity += req.Quantity
		if err := uc.repo.SaveCartItem(ctx, existingItem); err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		newItem := &entities.CartItem{CartID: userCart.ID, ProductID: req.ProductID, Quantity: req.Quantity}
		if err := uc.repo.CreateCartItem(ctx, newItem); err != nil {
			return nil, err
		}
	} else {
		return nil, err
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
