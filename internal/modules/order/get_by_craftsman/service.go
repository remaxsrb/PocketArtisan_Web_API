package get_by_craftsman

import (
	"PocketArtisan/internal/modules/order"
	"PocketArtisan/internal/modules/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  order.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: order.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req GetAllRequest) (GetAllResponse, error) {
	const maxLimit = 100
	const defaultLimit = 20
	const cacheTTL = 3 * time.Minute

	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		req.Limit = maxLimit
	}

	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "orders")
	cacheKey := fmt.Sprintf("orders:craftsman:v:%d:%d:skip:%d:limit:%d", cacheVersion, req.CraftsmanID, req.Skip, req.Limit)

	cachedData, err := uc.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedResp GetAllResponse
		if jsonErr := json.Unmarshal([]byte(cachedData), &cachedResp); jsonErr == nil {
			return cachedResp, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error: %v\n", err)
	}

	total, err := uc.repo.CountByCraftsman(ctx, req.CraftsmanID)
	if err != nil {
		return GetAllResponse{}, err
	}

	raw, err := uc.repo.ListByCraftsman(ctx, req.CraftsmanID, req.Skip, req.Limit)
	if err != nil {
		return GetAllResponse{}, err
	}

	orderList := make([]*order.OrderResponse, 0, len(raw))
	for _, o := range raw {
		orderList = append(orderList, &order.OrderResponse{
			OrderID:        o.ID,
			CustomerID:     o.CustomerID,
			OrderDate:      o.CreatedAt,
			CompletionDate: o.CompletedAt,
			URL:            o.URL,
			Status:         o.Status,
		})
	}

	countRows, err := uc.repo.ShippedCategoryCountsByCraftsman(ctx, req.CraftsmanID)
	if err != nil {
		return GetAllResponse{}, err
	}

	categoryCounts := make([]CategoryOrderCount, 0, len(countRows))
	for _, row := range countRows {
		categoryCounts = append(categoryCounts, CategoryOrderCount{Category: row.Category, Count: row.Count})
	}

	resp := GetAllResponse{
		Orders:         orderList,
		Total:          total,
		Page:           (req.Skip / req.Limit) + 1,
		CategoryCounts: categoryCounts,
	}

	jsonData, marshalErr := json.Marshal(resp)
	if marshalErr == nil {
		_ = uc.cache.Set(ctx, cacheKey, jsonData, cacheTTL).Err()
	}

	return resp, nil
}
