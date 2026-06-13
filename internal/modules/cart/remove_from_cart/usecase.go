package removefromcart

import (
	"PocketArtisan/internal/entities"
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

func (uc *UseCase) Execute(ctx context.Context, req RemoveFromCartRequest) (*cart.CartItemResponse, error) {
	
	result := uc.db.WithContext(ctx).
		Where("cart_id = ? AND product_id = ?", req.CartID, req.ProductID).
		Delete(&entities.CartItem{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("item not found in cart")
	}

	var item entities.CartItem
	if err := uc.db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", req.CartID, req.ProductID).Find(&item).Error; err != nil {
		return nil, err
	}

	var response cart.CartItemResponse
	response.ProductID = item.ProductID
	response.Quantity = item.Quantity

	return &response, nil
}
