package reject

import (
	"PocketArtisan/internal/entities"
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

func (uc *Service) Execute(ctx context.Context, req Request) error {
	var ca entities.CraftsmanApplication

	if err := uc.db.WithContext(ctx).Where("id = ?", req.ApplicationID).First(&ca).Error; err != nil {
		return errors.New("application not found")
	}

	ca.Status = entities.StatusRejected

	if err := uc.db.Save(&ca).Error; err != nil {
		return err
	}

	return nil
}
