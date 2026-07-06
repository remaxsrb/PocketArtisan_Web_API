package get_monthly_shipped

import (
	"PocketArtisan/internal/entities"
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
	db          *gorm.DB
	cache       *redis.Client
	timeService timeutil.Service
}

func NewService(db *gorm.DB, cache *redis.Client, timeService timeutil.Service) *Service {
	return &Service{db: db, cache: cache, timeService: timeService}
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

	query := uc.db.WithContext(ctx).
		Model(&entities.Order{}).
		Select("to_char(date_trunc('month', completed_at), 'YYYY-MM') as month, COUNT(*) as count").
		Where("status = ?", entities.OrderShipped)

	if from != nil {
		query = query.Where("completed_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("completed_at < ?", *to)
	}

	results := make([]MonthlyShippedCount, 0)
	if err := query.Group("month").Order("month").Scan(&results).Error; err != nil {
		return nil, err
	}

	if data, err := json.Marshal(results); err == nil {
		_ = uc.cache.Set(ctx, cacheKey, data, cacheTTL).Err()
	}

	return results, nil
}