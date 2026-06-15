package cart

import (
	"PocketArtisan/internal/entities"
	"context"
	"errors"

	"gorm.io/gorm"
)

// CartReader is the cart module contract used by other modules.
type CartReader interface {
	GetUserCart(ctx context.Context, userID uint64) (*entities.Cart, error)
}

type gormCartReader struct {
	db *gorm.DB
}

func NewCartReader(db *gorm.DB) CartReader {
	return &gormCartReader{db: db}
}

func (r *gormCartReader) GetUserCart(ctx context.Context, userID uint64) (*entities.Cart, error) {
	var userCart entities.Cart
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Images").
		Where("user_id = ?", userID).
		First(&userCart).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &entities.Cart{}, nil
		}
		return nil, err
	}

	return &userCart, nil
}
