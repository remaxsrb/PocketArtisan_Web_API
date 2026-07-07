package product

import (
	"PocketArtisan/internal/entities"
	"context"

	"gorm.io/gorm"
)

// Service is the cross-module contract used by product_categories and other
// modules that only need craftsman lookups.
type Service interface {
	GetCraftsmanIDByUsername(ctx context.Context, username string) (uint64, error)
	GetCraftsmanByUsername(ctx context.Context, username string) (*entities.Craftsman, error)
}

type serviceImpl struct {
	repo Repository
}

func NewService(db *gorm.DB) Service {
	return &serviceImpl{repo: NewGormRepository(db)}
}

func (s *serviceImpl) GetCraftsmanIDByUsername(ctx context.Context, username string) (uint64, error) {
	return s.repo.FindCraftsmanIDByUsername(ctx, username)
}

func (s *serviceImpl) GetCraftsmanByUsername(ctx context.Context, username string) (*entities.Craftsman, error) {
	return s.repo.FindCraftsmanByUsername(ctx, username)
}
