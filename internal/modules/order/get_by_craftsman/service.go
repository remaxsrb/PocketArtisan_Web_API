package get_by_craftsman

import (
	"PocketArtisan/internal/entities"
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

	var total int64
	uc.db.WithContext(ctx).Model(&entities.Order{}).Where("craftsman_id = ?", req.CraftsmanID).Count(&total)

	raw := make([]*entities.Order, 0, req.Limit)
	uc.db.WithContext(ctx).
		Where("craftsman_id = ?", req.CraftsmanID).
		Offset(req.Skip).
		Limit(req.Limit).
		Order("created_at desc").
		Find(&raw)

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

	categoryCounts := make([]CategoryOrderCount, 0)
	uc.db.WithContext(ctx).
		Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Joins("JOIN product_categories ON product_categories.id = products.category_id").
		Select("product_categories.name as category, COUNT(DISTINCT orders.id) as count").
		Where("orders.craftsman_id = ? AND orders.status = ?", req.CraftsmanID, entities.OrderShipped).
		Group("product_categories.name").
		Order("product_categories.name").
		Scan(&categoryCounts)

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
