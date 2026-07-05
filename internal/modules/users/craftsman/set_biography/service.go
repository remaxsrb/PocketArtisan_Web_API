package set_biography

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
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

func (uc *Service) Execute(ctx context.Context, req SetBiographyRequest) error {
	var craftsman entities.Craftsman

	if err := uc.db.WithContext(ctx).
		Where("id = ?", req.CraftsmanID).
		First(&craftsman).Error; err != nil {
		return errors.New("craftsman not found")
	}

	craftsman.Biography = req.Biography

	if err := uc.db.WithContext(ctx).Save(&craftsman).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "craftsmen")

	return nil
}