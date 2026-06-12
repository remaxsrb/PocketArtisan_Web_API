package delete

import (
	"PocketArtisan/internal/modules/crafts"
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

func (uc *UseCase) Execute(ctx context.Context, req DeleteCraftRequest) error {

	var c crafts.Craft
	if err := uc.db.WithContext(ctx).Where("name = ?", req.Name).First(&c).Error; err != nil {
		return errors.New("craft does not exist")
	}

	if err := uc.db.WithContext(ctx).Delete(&c).Error; err != nil {
		return err
	}

	return nil

}
