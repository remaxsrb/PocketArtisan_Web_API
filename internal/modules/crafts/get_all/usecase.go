package get_all

import (
	"PocketArtisan/internal/modules/crafts"
	"context"

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

func (uc *UseCase) Execute(ctx context.Context) ([]crafts.Craft, error) {

	var craftsList []crafts.Craft
	if err := uc.db.WithContext(ctx).Find(&craftsList).Error; err != nil {
		return nil, err
	}

	return craftsList, nil

}