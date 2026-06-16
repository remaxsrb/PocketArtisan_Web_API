package get_by_craftsman

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/order"
	"PocketArtisan/internal/modules/utils"
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
	db    *gorm.DB
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{db: db, cache: cache}
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

	craftsmanID, err := strconv.ParseUint(req.CraftsmanID, 10, 64)
	if err != nil {
		return GetAllResponse{}, fmt.Errorf("invalid craftsman_id: %w", err)
	}

	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "orders")
	cacheKey := fmt.Sprintf("orders:craftsman:v:%d:%d:skip:%d:limit:%d", cacheVersion, craftsmanID, req.Skip, req.Limit)

	cachedData, err := uc.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedResp GetAllResponse
		if jsonErr := json.Unmarshal([]byte(cachedData), &cachedResp); jsonErr == nil {
			return cachedResp, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error: %v\n", err)
	}

	var total int64
	uc.db.WithContext(ctx).Model(&entities.Order{}).Where("craftsman_id = ?", craftsmanID).Count(&total)

	raw := make([]*entities.Order, 0, req.Limit)
	uc.db.WithContext(ctx).
		Where("craftsman_id = ?", craftsmanID).
		Offset(req.Skip).
		Limit(req.Limit).
		Order("created_at desc").
		Find(&raw)

	orderList := make([]*order.OrderResponse, 0, len(raw))
	for _, o := range raw {
		orderList = append(orderList, &order.OrderResponse{
			OrderID:        o.ID,
			OrderDate:      o.CreatedAt,
			CompletionDate: o.CompletedAt,
			URL:            o.URL,
			Status:         o.Status,
		})
	}

	resp := GetAllResponse{
		Orders: orderList,
		Total:  total,
		Page:   (req.Skip / req.Limit) + 1,
	}

	jsonData, marshalErr := json.Marshal(resp)
	if marshalErr == nil {
		_ = uc.cache.Set(ctx, cacheKey, jsonData, cacheTTL).Err()
	}

	return resp, nil
}
