package removefromcart

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

func (uc *UseCase) Execute(ctx context.Context, req RemoveFromCartRequest) (*RemoveFromCartResponse, error) {
	var userCart cart.Cart
	err := uc.db.WithContext(ctx).
		Where("user_id = ?", req.UserID).
		First(&userCart).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cart not found")
		}
		return nil, err
	}

	result := uc.db.WithContext(ctx).
		Where("cart_id = ? AND product_id = ?", userCart.ID, req.ProductID).
		Delete(&cart.CartItem{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("item not found in cart")
	}

	var items []cart.CartItem
	if err := uc.db.WithContext(ctx).Where("cart_id = ?", userCart.ID).Find(&items).Error; err != nil {
		return nil, err
	}

	var response RemoveFromCartResponse
	for _, item := range items {
		response.CartItems = append(response.CartItems, cart.CartItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return &response, nil
}
