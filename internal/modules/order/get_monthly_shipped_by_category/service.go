package get_monthly_shipped_by_category

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
	"PocketArtisan/internal/modules/utils/timeutil"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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

func (uc *Service) Execute(ctx context.Context, req MonthlyShippedByCategoryRequest) ([]MonthlyCategoryCount, error) {
	const cacheTTL = 5 * time.Minute

	craftsmanID, err := strconv.ParseUint(req.CraftsmanID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid craftsman_id: %w", err)
	}

	from, to, err := uc.timeService.ParseDateRange(req.From, req.To)
	if err != nil {
		return nil, err
	}

	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "orders")
	cacheKey := fmt.Sprintf("orders:stats:monthly:craftsman:v:%d:%d:%s:%s", cacheVersion, craftsmanID, req.From, req.To)

	if cached, err := uc.cache.Get(ctx, cacheKey).Result(); err == nil {
		cachedResp := make([]MonthlyCategoryCount, 0)
		if jsonErr := json.Unmarshal([]byte(cached), &cachedResp); jsonErr == nil {
			return cachedResp, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error: %v\n", err)
	}

	query := uc.db.WithContext(ctx).
		Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Joins("JOIN product_categories ON product_categories.id = products.category_id").
		Select(`
			to_char(date_trunc('month', orders.completed_at), 'YYYY-MM') as month,
			product_categories.name as category,
			COUNT(DISTINCT orders.id) as count
		`).
		Where("orders.craftsman_id = ? AND orders.status = ?", craftsmanID, entities.OrderShipped)

	if from != nil {
		query = query.Where("orders.completed_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("orders.completed_at < ?", *to)
	}

	results := make([]MonthlyCategoryCount, 0)
	if err := query.Group("month, product_categories.name").Order("month").Scan(&results).Error; err != nil {
		return nil, err
	}

	if data, err := json.Marshal(results); err == nil {
		_ = uc.cache.Set(ctx, cacheKey, data, cacheTTL).Err()
	}

	return results, nil
}