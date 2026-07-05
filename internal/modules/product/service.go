package product

import (
	"PocketArtisan/internal/entities"
	"context"
	"errors"

	"gorm.io/gorm"
)

type Service interface {
	GetCraftsmanIDByUsername(ctx context.Context, username string) (uint64, error)
	GetCraftsmanByUsername(ctx context.Context, username string) (*entities.Craftsman, error)
}

type gormService struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &gormService{db: db}
}

func (s *gormService) GetCraftsmanIDByUsername(ctx context.Context, username string) (uint64, error) {
	var craftsman entities.Craftsman
	if err := s.db.WithContext(ctx).
		Joins("JOIN users ON users.id = craftsmen.user_id").
		Where("users.username = ?", username).
		First(&craftsman).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("craftsman not found")
		}
		return 0, err
	}
	return craftsman.ID, nil
}

func (s *gormService) GetCraftsmanByUsername(ctx context.Context, username string) (*entities.Craftsman, error) {
	var craftsman entities.Craftsman
	if err := s.db.WithContext(ctx).
		Joins("JOIN users ON users.id = craftsmen.user_id").
		Where("users.username = ?", username).
		First(&craftsman).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("craftsman not found")
		}
		return nil, err
	}
	return &craftsman, nil
}
