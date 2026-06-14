package addtocart

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

func (uc *UseCase) Execute(ctx context.Context, req AddToCartRequest) (*cart.CartResponse, error) {
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

	var product_price = 0.0

	err = uc.db.WithContext(ctx).
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

	userCart.Total += product_price * float64(req.Quantity)
	uc.db.WithContext(ctx).Save(&userCart)

	response.Cart = userCart

	return &response, nil
}
