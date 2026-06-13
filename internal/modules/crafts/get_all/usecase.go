package get_all

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

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

func (uc *UseCase) Execute(ctx context.Context) ([]entities.Craft, error) {
	const cacheTTL = 5 * time.Minute

	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "crafts")
	cacheKey := fmt.Sprintf("crafts:all:v:%d", cacheVersion)

	if cached, err := uc.cache.Get(ctx, cacheKey).Result(); err == nil {
		var craftsList []entities.Craft
		if err := json.Unmarshal([]byte(cached), &craftsList); err == nil {
			return craftsList, nil
		}
	}

	var craftsList []entities.Craft
	if err := uc.db.WithContext(ctx).Find(&craftsList).Error; err != nil {
		return nil, err
	}

	if data, err := json.Marshal(craftsList); err == nil {
		_ = uc.cache.Set(ctx, cacheKey, data, cacheTTL).Err()
	}

	return craftsList, nil

}
