package addtocart

import (
	"PocketArtisan/internal/entities"
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

	var existingItem entities.CartItem
	err := uc.db.WithContext(ctx).
		Where("cart_id = ? AND product_id = ?", req.CartID, req.ProductID).
		First(&existingItem).
		Error

	if err == nil {
		existingItem.Quantity += req.Quantity
		if err := uc.db.WithContext(ctx).Save(&existingItem).Error; err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		newItem := entities.CartItem{
			CartID:    req.CartID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if err := uc.db.WithContext(ctx).Create(&newItem).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	var items []entities.CartItem
	if err := uc.db.WithContext(ctx).Where("cart_id = ?", req.CartID).Find(&items).Error; err != nil {
		return nil, err
	}

	var response AddToCartResponse
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

	response.Cart = userCart

	return &response, nil
}
