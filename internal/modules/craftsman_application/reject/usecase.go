package reject

import (
	"PocketArtisan/internal/modules/craftsman_application"
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

func (uc *UseCase) Execute(ctx context.Context, req Request) error {
	var ca craftsman_application.CraftsmanApplication

	if err := uc.db.WithContext(ctx).Where("id = ?", req.ApplicationID).First(&ca).Error; err != nil {
		return errors.New("application not found")
	}

	ca.Status = craftsman_application.StatusRejected

	if err := uc.db.Save(&ca).Error; err != nil {
		return err
	}

	return nil
}
