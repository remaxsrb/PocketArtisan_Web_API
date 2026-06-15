package removefromcart

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/cart"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{db: db, cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req RemoveFromCartRequest) (*cart.CartResponse, error) {

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

	var product_price = 0.0

	err := uc.db.WithContext(ctx).
		Model(entities.Product{}).
		Select("price").
		Where("id = ?", req.ProductID).
		Scan(&product_price).Error
	if err != nil {
		return nil, err
	}

	var response cart.CartResponse
	var userCart entities.Cart
	cartErr := uc.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Images").
		Where("user_id = ?", req.CartID).
		First(&userCart).
		Error

	if cartErr != nil && !errors.Is(cartErr, gorm.ErrRecordNotFound) {
		return nil, cartErr
	}

	userCart.Total -= product_price * float64(req.Quantity)
	uc.db.WithContext(ctx).Save(&userCart)
	response.Cart = userCart

	return &response, nil
}
