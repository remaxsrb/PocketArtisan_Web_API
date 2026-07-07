package get_monthly_shipped

import (
	ordermod "PocketArtisan/internal/modules/order"
	"PocketArtisan/internal/modules/utils"
	"PocketArtisan/internal/modules/utils/timeutil"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo        ordermod.Repository
	cache       *redis.Client
	timeService timeutil.Service
}

func NewService(db *gorm.DB, cache *redis.Client, timeService timeutil.Service) *Service {
	return &Service{repo: ordermod.NewGormRepository(db), cache: cache, timeService: timeService}
}

func (uc *Service) Execute(ctx context.Context, req MonthlyShippedRequest) ([]MonthlyShippedCount, error) {
	const cacheTTL = 5 * time.Minute

	from, to, err := uc.timeService.ParseDateRange(req.From, req.To)
	if err != nil {
		return nil, err
	}

	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "orders")
	cacheKey := fmt.Sprintf("orders:stats:monthly:v:%d:%s:%s", cacheVersion, req.From, req.To)

	if cached, err := uc.cache.Get(ctx, cacheKey).Result(); err == nil {
		cachedResp := make([]MonthlyShippedCount, 0)
		if jsonErr := json.Unmarshal([]byte(cached), &cachedResp); jsonErr == nil {
			return cachedResp, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error: %v\n", err)
	}

	rows, err := uc.repo.MonthlyShippedCounts(ctx, from, to)
	if err != nil {
		return nil, err
	}

	results := make([]MonthlyShippedCount, 0, len(rows))
	for _, row := range rows {
		results = append(results, MonthlyShippedCount{Month: row.Month, Count: row.Count})
	}

	if data, err := json.Marshal(results); err == nil {
		_ = uc.cache.Set(ctx, cacheKey, data, cacheTTL).Err()
	}

	return results, nil
}
