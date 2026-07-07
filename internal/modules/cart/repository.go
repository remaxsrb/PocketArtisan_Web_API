package cart

import (
	"PocketArtisan/internal/entities"
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository covers all cart data access operations.
// It supersedes the narrower CartReader interface, which is kept for
// cross-module use (modules that only need GetUserCart).
type Repository interface {
	GetUserCart(ctx context.Context, userID uint64) (*entities.Cart, error)
	FindCartItem(ctx context.Context, cartID, productID uint64) (*entities.CartItem, error)
	GetProductPrice(ctx context.Context, productID uint64) (float64, error)
	SaveCartItem(ctx context.Context, item *entities.CartItem) error
	CreateCartItem(ctx context.Context, item *entities.CartItem) error
	DeleteCartItem(ctx context.Context, cartID, productID uint64) (int64, error)
	SaveCart(ctx context.Context, cart *entities.Cart) error
	ClearCartItems(ctx context.Context, cartID uint64) error
	UpdateCartTotal(ctx context.Context, cartID uint64, total float64) error
	RefreshCartTotal(ctx context.Context, cartID uint64) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) GetUserCart(ctx context.Context, userID uint64) (*entities.Cart, error) {
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

func (r *GormRepository) FindCartItem(ctx context.Context, cartID, productID uint64) (*entities.CartItem, error) {
	var item entities.CartItem
	err := r.db.WithContext(ctx).
		Where("cart_id = ? AND product_id = ?", cartID, productID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) GetProductPrice(ctx context.Context, productID uint64) (float64, error) {
	var price float64
	err := r.db.WithContext(ctx).
		Model(entities.Product{}).
		Select("price").
		Where("id = ?", productID).
		Scan(&price).Error
	return price, err
}

func (r *GormRepository) SaveCartItem(ctx context.Context, item *entities.CartItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *GormRepository) CreateCartItem(ctx context.Context, item *entities.CartItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GormRepository) DeleteCartItem(ctx context.Context, cartID, productID uint64) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("cart_id = ? AND product_id = ?", cartID, productID).
		Delete(&entities.CartItem{})
	return result.RowsAffected, result.Error
}

func (r *GormRepository) SaveCart(ctx context.Context, cart *entities.Cart) error {
	return r.db.WithContext(ctx).Save(cart).Error
}

func (r *GormRepository) ClearCartItems(ctx context.Context, cartID uint64) error {
	return r.db.WithContext(ctx).
		Where("cart_id = ?", cartID).
		Delete(&entities.CartItem{}).Error
}

func (r *GormRepository) UpdateCartTotal(ctx context.Context, cartID uint64, total float64) error {
	return r.db.WithContext(ctx).
		Model(&entities.Cart{}).
		Where("id = ?", cartID).
		Update("total", total).Error
}

func (r *GormRepository) RefreshCartTotal(ctx context.Context, cartID uint64) error {
	var total float64
	err := r.db.WithContext(ctx).
		Table("cart_items").
		Joins("JOIN products ON products.id = cart_items.product_id").
		Where("cart_items.cart_id = ?", cartID).
		Select("COALESCE(SUM(cart_items.quantity * products.price), 0)").
		Scan(&total).Error
	if err != nil {
		return err
	}
	return r.UpdateCartTotal(ctx, cartID, total)
}
