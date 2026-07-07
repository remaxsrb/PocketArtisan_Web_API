package cart

import (
	"PocketArtisan/internal/entities"
	"context"

	"gorm.io/gorm"
)

// CartReader is the narrow cross-module interface for reading a user's cart.
// Modules that only need this one method should accept CartReader, not Repository.
type CartReader interface {
	GetUserCart(ctx context.Context, userID uint64) (*entities.Cart, error)
}

// NewCartReader returns a CartReader backed by the full GormRepository.
// Existing callers (login, checkout) remain unchanged.
func NewCartReader(db *gorm.DB) CartReader {
	return NewGormRepository(db)
}
