package addtocart

import (
	"PocketArtisan/internal/modules/cart"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UseCase struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewUseCase(db *gorm.DB, cache *redis.Client) *UseCase {
	return &UseCase{db: db, cache: cache}
}

func (uc *UseCase) Execute(ctx context.Context, req AddToCartRequest) (*AddToCartResponse, error) {
	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	var userCart cart.Cart
	err := uc.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", req.UserID).
		First(&userCart, cart.Cart{UserID: req.UserID}).
		Error
	if err != nil {
		return nil, err
	}

	var existingItem cart.CartItem
	err = uc.db.WithContext(ctx).
		Where("cart_id = ? AND product_id = ?", userCart.ID, req.ProductID).
		First(&existingItem).
		Error

	if err == nil {
		existingItem.Quantity += req.Quantity
		if err := uc.db.WithContext(ctx).Save(&existingItem).Error; err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		newItem := cart.CartItem{
			CartID:    userCart.ID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if err := uc.db.WithContext(ctx).Create(&newItem).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	var items []cart.CartItem
	if err := uc.db.WithContext(ctx).Where("cart_id = ?", userCart.ID).Find(&items).Error; err != nil {
		return nil, err
	}

	var response AddToCartResponse
	for _, item := range items {
		response.CartItems = append(response.CartItems, cart.CartItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return &response, nil
}
